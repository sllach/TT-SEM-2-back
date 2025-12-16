package material

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateMaterial maneja la creación de un material
func CreateMaterial(c *gin.Context) {
	// 1. Obtener GoogleID del usuario autenticado
	googleIDAny, exists := c.Get("google_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Datos de usuario incompletos"})
		return
	}
	googleID := googleIDAny.(string)

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// 2. Parsear form-data (max 32MB para uploads)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parseando form-data: " + err.Error()})
		return
	}

	// 3. Extraer y Validar campos básicos
	nombre := c.PostForm("nombre")
	if nombre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'nombre' es requerido"})
		return
	}
	descripcion := c.PostForm("descripcion")

	// 4. Parsear JSONs Complejos

	// Herramientas (Array de Strings)
	var herramientas models.StringArray
	if str := c.PostForm("herramientas"); str != "" {
		if err := json.Unmarshal([]byte(str), &herramientas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de herramientas inválido"})
			return
		}
	}

	// Composición (Array de Objetos: Elemento + Cantidad)
	var composicion models.JSONComponentes
	if str := c.PostForm("composicion"); str != "" {
		if err := json.Unmarshal([]byte(str), &composicion); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de composición inválido"})
			return
		}
	}

	// Propiedades Mecánicas (Nombre + Valor + Unidad)
	var propMecanicas models.JSONMecanicas
	if str := c.PostForm("prop_mecanicas"); str != "" {
		if err := json.Unmarshal([]byte(str), &propMecanicas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de prop. mecánicas inválido"})
			return
		}
	}

	// Propiedades Perceptivas (Nombre + Valor)
	var propPerceptivas models.JSONGenerales
	if str := c.PostForm("prop_perceptivas"); str != "" {
		if err := json.Unmarshal([]byte(str), &propPerceptivas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de prop. perceptivas inválido"})
			return
		}
	}

	// Propiedades Emocionales (Nombre + Valor)
	var propEmocionales models.JSONGenerales
	if str := c.PostForm("prop_emocionales"); str != "" {
		if err := json.Unmarshal([]byte(str), &propEmocionales); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de prop. emocionales inválido"})
			return
		}
	}

	// Derivado De (UUID opcional)
	var derivadoDe uuid.UUID
	if str := c.PostForm("derivado_de"); str != "" {
		parsedUUID, err := uuid.Parse(str)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "UUID de derivado_de inválido"})
			return
		}
		derivadoDe = parsedUUID
	}

	// 5. Crear el Objeto Material
	material := models.Material{
		ID:                     uuid.New(),
		Nombre:                 nombre,
		Descripcion:            descripcion,
		Herramientas:           herramientas,
		Composicion:            composicion,
		PropiedadesMecanicas:   propMecanicas,
		PropiedadesPerceptivas: propPerceptivas,
		PropiedadesEmocionales: propEmocionales,
		DerivadoDe:             derivadoDe,
		CreadorID:              googleID,
		Estado:                 false, // Pendiente de aprobación
	}

	// Guardar Material (Esto guarda automáticamente los JSONs en las columnas jsonb)
	if err := db.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando material en BD: " + err.Error()})
		return
	}

	// 6. Guardar Colaboradores
	colaboradoresStr := c.PostForm("colaboradores")
	if colaboradoresStr == "" {
		colaboradoresStr = c.PostForm("colaboradores_material")
	}

	if colaboradoresStr != "" {
		var emails []string
		if err := json.Unmarshal([]byte(colaboradoresStr), &emails); err == nil && len(emails) > 0 {
			var usuarios []models.Usuario
			// Buscar usuarios por email
			if err := db.Where("email IN ?", emails).Find(&usuarios).Error; err == nil {
				// Insertar relaciones manualmente
				for _, u := range usuarios {
					link := models.ColaboradorMaterial{
						MaterialID: material.ID,
						UsuarioID:  u.GoogleID,
					}
					// Usamos db.Create directo sobre la estructura, GORM usará la tabla definida en TableName()
					if err := db.Create(&link).Error; err != nil {
						log.Printf("⚠️ Error asociando colaborador %s: %v", u.Email, err)
					}
				}
			}
		}
	}

	// 7. Guardar Galería
	var galeriaCaptions []string
	if str := c.PostForm("galeria_captions"); str != "" {
		json.Unmarshal([]byte(str), &galeriaCaptions)
	}

	files := c.Request.MultipartForm.File["galeria_images[]"]
	for i, fileHeader := range files {
		safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
		filePath := fmt.Sprintf("materials/%s/%s", material.ID.String(), safeFilename)

		url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
		if err != nil {
			log.Printf("Error subiendo imagen galería: %v", err)
			continue
		}

		caption := ""
		if i < len(galeriaCaptions) {
			caption = galeriaCaptions[i]
		}

		db.Create(&models.GaleriaMaterial{
			MaterialID: material.ID,
			URLImagen:  url,
			Caption:    caption,
		})
	}

	// 8. Guardar Pasos
	pasosStr := c.PostForm("pasos")
	if pasosStr != "" {
		var pasos []struct {
			OrdenPaso   int    `json:"orden_paso"`
			Descripcion string `json:"descripcion"`
		}
		if err := json.Unmarshal([]byte(pasosStr), &pasos); err == nil {
			for i, p := range pasos {
				pasoModel := models.PasoMaterial{
					MaterialID:  material.ID,
					OrdenPaso:   p.OrdenPaso,
					Descripcion: p.Descripcion,
				}

				// Upload Imagen Paso
				fileKeyImg := fmt.Sprintf("paso_images[%d]", i)
				if headers := c.Request.MultipartForm.File[fileKeyImg]; len(headers) > 0 {
					safeName := strings.ReplaceAll(headers[0].Filename, " ", "_")
					path := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeName)
					if url, err := database.SubirAStorageSupabase(headers[0], "pasos-bucket", path); err == nil {
						pasoModel.URLImagen = url
					}
				}

				// Upload Video Paso
				fileKeyVid := fmt.Sprintf("paso_videos[%d]", i)
				if headers := c.Request.MultipartForm.File[fileKeyVid]; len(headers) > 0 {
					safeName := strings.ReplaceAll(headers[0].Filename, " ", "_")
					path := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeName)
					if url, err := database.SubirAStorageSupabase(headers[0], "pasos-bucket", path); err == nil {
						pasoModel.URLVideo = url
					}
				}

				db.Create(&pasoModel)
			}
		}
	}

	// 9. Recargar y Responder
	db.Preload("Creador").Preload("Colaboradores").Preload("Galeria").Preload("Pasos").Find(&material)

	notificarAdmins(material.ID, material.Nombre, material.CreadorID)
	c.JSON(http.StatusCreated, material)
}

// Función auxiliar de notificaciones
func notificarAdmins(matID uuid.UUID, matNombre string, creadorID string) {
	go func() {
		db, _ := database.GetDB()
		var creador models.Usuario
		if err := db.Where("google_id = ?", creadorID).First(&creador).Error; err != nil {
			creador.Nombre = "Usuario"
			creador.Email = creadorID
		}
		var admins []models.Usuario
		db.Where("rol = ?", "administrador").Find(&admins)

		mensaje := fmt.Sprintf("El usuario %s (%s) ha subido '%s'. Requiere revisión.", creador.Nombre, creador.Email, matNombre)
		for _, admin := range admins {
			db.Create(&models.Notificacion{
				UsuarioID:  admin.GoogleID,
				MaterialID: &matID,
				Titulo:     "Nuevo Material Pendiente",
				Mensaje:    mensaje,
				Tipo:       "info",
				Link:       "/admin",
			})
		}
	}()
}
