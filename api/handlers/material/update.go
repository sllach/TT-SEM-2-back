package material

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateMaterial maneja la actualización de un material
func UpdateMaterial(c *gin.Context) {
	// 1. Obtener usuario autenticado
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Datos de usuario incompletos"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// 2. Verificar existencia y permisos
	var material models.Material
	// Cargamos Pasos y Galeria para poder editarlos.
	if err := db.Preload("Pasos").Preload("Galeria").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	isAdmin := middleware.IsAdmin(c)
	isOwner := material.CreadorID == googleID

	if !isAdmin && !isOwner {
		c.JSON(http.StatusForbidden, gin.H{
			"error":  "No tienes permiso",
			"detail": "Solo puedes editar tus propios materiales",
		})
		return
	}

	// 3. Parsear Multipart Form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parseando form-data: " + err.Error()})
		return
	}

	// 4. Actualizar Campos de Texto y JSONs

	// Textos Simples
	if val := c.PostForm("nombre"); val != "" {
		material.Nombre = val
	}
	if val := c.PostForm("descripcion"); val != "" {
		material.Descripcion = val
	}

	// Herramientas (Array String)
	if str := c.PostForm("herramientas"); str != "" {
		var h models.StringArray
		if err := json.Unmarshal([]byte(str), &h); err == nil {
			material.Herramientas = h
		}
	}

	// Composición (JSONComponentes)
	if str := c.PostForm("composicion"); str != "" {
		var comp models.JSONComponentes
		if err := json.Unmarshal([]byte(str), &comp); err == nil {
			material.Composicion = comp
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON Composición inválido"})
			return
		}
	}

	// Propiedades Mecánicas (JSONMecanicas)
	if str := c.PostForm("prop_mecanicas"); str != "" {
		var pm models.JSONMecanicas
		if err := json.Unmarshal([]byte(str), &pm); err == nil {
			material.PropiedadesMecanicas = pm
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON Prop. Mecánicas inválido"})
			return
		}
	}

	// Propiedades Perceptivas (JSONGenerales)
	if str := c.PostForm("prop_perceptivas"); str != "" {
		var pp models.JSONGenerales
		if err := json.Unmarshal([]byte(str), &pp); err == nil {
			material.PropiedadesPerceptivas = pp
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON Prop. Perceptivas inválido"})
			return
		}
	}

	// Propiedades Emocionales (Nuevo Tipo: JSONGenerales)
	if str := c.PostForm("prop_emocionales"); str != "" {
		var pe models.JSONGenerales
		if err := json.Unmarshal([]byte(str), &pe); err == nil {
			material.PropiedadesEmocionales = pe
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON Prop. Emocionales inválido"})
			return
		}
	}

	// Derivado De
	if str := c.PostForm("derivado_de"); str != "" {
		if uid, err := uuid.Parse(str); err == nil {
			material.DerivadoDe = uid
		}
	}

	// Resetear estado al editar
	material.Estado = false

	// Guardar Cambios en Material (Actualiza columnas JSON automáticamente)
	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando actualización: " + err.Error()})
		return
	}

	// 5. Actualizar Colaboradores (Manual y Seguro)
	colaboradoresStr := c.PostForm("colaboradores")
	if colaboradoresStr != "" {
		// A. Borrar anteriores usando nombre de tabla explícito
		if err := db.Table("material_colaboradores").Where("material_id = ?", material.ID).Delete(nil).Error; err != nil {
			log.Printf("⚠️ Error limpiando colaboradores antiguos: %v", err)
		}

		// B. Insertar nuevos
		var emails []string
		if err := json.Unmarshal([]byte(colaboradoresStr), &emails); err == nil && len(emails) > 0 {
			var usuarios []models.Usuario
			db.Where("email IN ?", emails).Find(&usuarios)

			for _, u := range usuarios {
				link := models.ColaboradorMaterial{
					MaterialID: material.ID,
					UsuarioID:  u.GoogleID,
				}
				// Insertar uno por uno
				if err := db.Create(&link).Error; err != nil {
					log.Printf("⚠️ Error agregando colaborador %s: %v", u.Email, err)
				}
			}
		}
	}

	// 6. Actualizar Galería
	galeriaCaptionsStr := c.PostForm("galeria_captions")
	var galeriaCaptions []string
	if galeriaCaptionsStr != "" {
		json.Unmarshal([]byte(galeriaCaptionsStr), &galeriaCaptions)
	}

	files := c.Request.MultipartForm.File["galeria_images[]"]
	if len(files) > 0 {
		// Si suben nuevas fotos, reemplazamos todo (Estrategia simple)
		db.Where("material_id = ?", material.ID).Delete(&models.GaleriaMaterial{})

		for i, fileHeader := range files {
			safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
			filePath := fmt.Sprintf("materials/%s/%s", material.ID.String(), safeFilename)

			url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
			if err != nil {
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
	} else if len(galeriaCaptions) > 0 {
		// Si solo actualizan textos de galería existente
		sort.Slice(material.Galeria, func(i, j int) bool {
			return material.Galeria[i].ID < material.Galeria[j].ID // Asume ID autoincremental o ordenable
		})
		for i, caption := range galeriaCaptions {
			if i < len(material.Galeria) {
				material.Galeria[i].Caption = caption
				db.Save(&material.Galeria[i])
			}
		}
	}

	// 7. Actualizar Pasos
	pasosStr := c.PostForm("pasos")
	if pasosStr != "" {
		var newPasos []struct {
			OrdenPaso   int    `json:"orden_paso"`
			Descripcion string `json:"descripcion"`
		}
		if err := json.Unmarshal([]byte(pasosStr), &newPasos); err == nil {
			// Mapa de pasos existentes
			existingPasos := make(map[int]models.PasoMaterial)
			for _, p := range material.Pasos {
				existingPasos[p.OrdenPaso] = p
			}

			for i, newPaso := range newPasos {
				var pasoModel models.PasoMaterial
				if exist, ok := existingPasos[newPaso.OrdenPaso]; ok {
					pasoModel = exist
					pasoModel.Descripcion = newPaso.Descripcion
				} else {
					pasoModel = models.PasoMaterial{
						MaterialID:  material.ID,
						OrdenPaso:   newPaso.OrdenPaso,
						Descripcion: newPaso.Descripcion,
					}
				}

				// Uploads (Imagen/Video) para este paso
				fileKeyImg := fmt.Sprintf("paso_images[%d]", i)
				if headers := c.Request.MultipartForm.File[fileKeyImg]; len(headers) > 0 {
					safeName := strings.ReplaceAll(headers[0].Filename, " ", "_")
					path := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), newPaso.OrdenPaso, safeName)
					if url, err := database.SubirAStorageSupabase(headers[0], "pasos-bucket", path); err == nil {
						pasoModel.URLImagen = url
					}
				}
				// (Repetir lógica para video si es necesario...)

				if pasoModel.ID == 0 {
					db.Create(&pasoModel)
				} else {
					db.Save(&pasoModel)
				}
			}

			// Limpiar pasos viejos
			newOrdens := make(map[int]bool)
			for _, np := range newPasos {
				newOrdens[np.OrdenPaso] = true
			}
			for orden, exist := range existingPasos {
				if !newOrdens[orden] {
					db.Delete(&exist)
				}
			}
		}
	}

	// 8. Respuesta Final
	db.Preload("Creador").Preload("Colaboradores").Preload("Galeria").Preload("Pasos").Find(&material)

	// Notificar
	var creador models.Usuario
	db.Where("google_id = ?", material.CreadorID).First(&creador)
	notificarUpdate(material.ID, material.Nombre, creador.Nombre)

	c.JSON(http.StatusOK, material)
}

// Función auxiliar para notificaciones
func notificarUpdate(matID uuid.UUID, matNombre string, creadorNombre string) {
	go func() {
		db, _ := database.GetDB()
		var admins []models.Usuario
		db.Where("rol = ?", "administrador").Find(&admins)

		mensaje := fmt.Sprintf("El usuario %s ha actualizado: '%s'. Requiere revisión.", creadorNombre, matNombre)
		for _, admin := range admins {
			db.Create(&models.Notificacion{
				UsuarioID:  admin.GoogleID,
				MaterialID: &matID,
				Titulo:     "Material Actualizado",
				Mensaje:    mensaje,
				Tipo:       "info",
				Link:       "/admin",
			})
		}
	}()
}
