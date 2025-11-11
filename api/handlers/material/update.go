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
// Los colaboradores solo pueden editar sus propios materiales
// Los administradores pueden editar cualquier material
func UpdateMaterial(c *gin.Context) {
	// Obtener GoogleID del usuario autenticado
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Datos de usuario incompletos"})
		return
	}

	db, err := database.OpenGormDB()
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

	// Verificar si el material existe y precargar relaciones necesarias
	var material models.Material
	if err := db.Preload("Pasos").Preload("Galeria").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// VERIFICACIÓN DE PERMISOS:
	// - Si es administrador: puede editar cualquier material
	// - Si es colaborador: solo puede editar sus propios materiales
	isAdmin := middleware.IsAdmin(c)
	isOwner := material.CreadorID == googleID

	if !isAdmin && !isOwner {
		c.JSON(http.StatusForbidden, gin.H{
			"error":  "No tienes permiso para editar este material",
			"detail": "Los colaboradores solo pueden editar sus propios materiales",
		})
		return
	}

	// Parsear form-data (max 32MB para uploads)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parseando form-data: " + err.Error()})
		return
	}

	// Extraer campos textuales
	nombre := c.PostForm("nombre")
	descripcion := c.PostForm("descripcion")
	herramientasStr := c.PostForm("herramientas")
	composicionStr := c.PostForm("composicion")
	derivadoDeStr := c.PostForm("derivado_de")
	propMecanicasStr := c.PostForm("prop_mecanicas")
	propPerceptivasStr := c.PostForm("prop_perceptivas")
	propEmocionalesStr := c.PostForm("prop_emocionales")
	colaboradoresStr := c.PostForm("colaboradores")
	pasosStr := c.PostForm("pasos")
	galeriaCaptionsStr := c.PostForm("galeria_captions")

	// Actualizar campos
	if nombre != "" {
		material.Nombre = nombre
	}
	if descripcion != "" {
		material.Descripcion = descripcion
	}
	if herramientasStr != "" {
		var herramientas models.StringArray
		if err := json.Unmarshal([]byte(herramientasStr), &herramientas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "herramientas inválido (debe ser JSON array)"})
			return
		}
		material.Herramientas = herramientas
	}
	if composicionStr != "" {
		var composicion models.StringArray
		if err := json.Unmarshal([]byte(composicionStr), &composicion); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "composicion inválido (debe ser JSON array)"})
			return
		}
		material.Composicion = composicion
	}
	if derivadoDeStr != "" {
		derivadoDe, err := uuid.Parse(derivadoDeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "derivado_de inválido (UUID)"})
			return
		}
		material.DerivadoDe = derivadoDe
	}

	// Guardar cambios en el material principal
	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando material: " + err.Error()})
		return
	}

	// Propiedades mecánicas
	if propMecanicasStr != "" {
		var propMecanicas models.PropiedadesMecanicas
		if err := json.Unmarshal([]byte(propMecanicasStr), &propMecanicas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prop_mecanicas inválido (JSON object)"})
			return
		}
		propMecanicas.MaterialID = material.ID
		if err := db.Debug().Save(&propMecanicas).Error; err != nil {
			log.Printf("Error actualizando prop_mecanicas: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando propiedades mecánicas: " + err.Error()})
			return
		}
	}

	// Propiedades perceptivas
	if propPerceptivasStr != "" {
		var propPerceptivas models.PropiedadesPerceptivas
		if err := json.Unmarshal([]byte(propPerceptivasStr), &propPerceptivas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prop_perceptivas inválido (JSON object)"})
			return
		}
		propPerceptivas.MaterialID = material.ID
		if err := db.Debug().Save(&propPerceptivas).Error; err != nil {
			log.Printf("Error actualizando prop_perceptivas: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando propiedades perceptivas: " + err.Error()})
			return
		}
	}

	// Propiedades emocionales
	if propEmocionalesStr != "" {
		var propEmocionales models.PropiedadesEmocionales
		if err := json.Unmarshal([]byte(propEmocionalesStr), &propEmocionales); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prop_emocionales inválido (JSON object)"})
			return
		}
		propEmocionales.MaterialID = material.ID
		if err := db.Debug().Save(&propEmocionales).Error; err != nil {
			log.Printf("Error actualizando prop_emocionales: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando propiedades emocionales: " + err.Error()})
			return
		}
	}

	// Colaboradores
	if colaboradoresStr != "" {
		if err := db.Where("material_id = ?", material.ID).Delete(&models.ColaboradorMaterial{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando colaboradores existentes: " + err.Error()})
			return
		}

		var colaboradores []string
		if err := json.Unmarshal([]byte(colaboradoresStr), &colaboradores); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "colaboradores inválido (JSON array)"})
			return
		}
		for _, userID := range colaboradores {
			colaborador := models.ColaboradorMaterial{
				MaterialID: material.ID,
				UsuarioID:  userID,
			}
			if err := db.Debug().Create(&colaborador).Error; err != nil {
				log.Printf("Error creando colaborador: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando colaborador: " + err.Error()})
				return
			}
		}
	}

	// Parsear captions para galería
	var galeriaCaptions []string
	if galeriaCaptionsStr != "" {
		if err := json.Unmarshal([]byte(galeriaCaptionsStr), &galeriaCaptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "galeria_captions inválido (JSON array)"})
			return
		}
	}

	// Manejar galería
	files := c.Request.MultipartForm.File["galeria_images[]"]
	if len(files) > 0 {
		if err := db.Where("material_id = ?", material.ID).Delete(&models.GaleriaMaterial{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando galería existente: " + err.Error()})
			return
		}

		for i, fileHeader := range files {
			safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
			filePath := fmt.Sprintf("materials/%s/%s", material.ID.String(), safeFilename)

			url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo a Supabase: " + err.Error()})
				return
			}

			caption := ""
			if i < len(galeriaCaptions) {
				caption = galeriaCaptions[i]
			}

			galeria := models.GaleriaMaterial{
				MaterialID: material.ID,
				URLImagen:  url,
				Caption:    caption,
			}
			if err := db.Debug().Create(&galeria).Error; err != nil {
				log.Printf("Error creando galeria: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando galería: " + err.Error()})
				return
			}
		}
	} else if len(galeriaCaptions) > 0 {
		sort.Slice(material.Galeria, func(i, j int) bool {
			return material.Galeria[i].ID < material.Galeria[j].ID
		})
		if len(galeriaCaptions) != len(material.Galeria) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Número de captions no coincide con galería existente"})
			return
		}
		for i, caption := range galeriaCaptions {
			material.Galeria[i].Caption = caption
			if err := db.Save(&material.Galeria[i]).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando caption de galería: " + err.Error()})
				return
			}
		}
	}

	// Manejar pasos (actualizar si proporcionado, manteniendo multimedia si no se cambian)
	if pasosStr != "" {
		var newPasos []struct {
			OrdenPaso   int    `json:"orden_paso"`
			Descripcion string `json:"descripcion"`
		}
		if err := json.Unmarshal([]byte(pasosStr), &newPasos); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pasos inválido (JSON array)"})
			return
		}

		// Mapa de pasos existentes por orden_paso
		existingPasos := make(map[int]models.PasoMaterial)
		for _, p := range material.Pasos {
			existingPasos[p.OrdenPaso] = p
		}

		// Procesar cada nuevo paso
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

			// Upload imagen si proporcionada
			fileKey := fmt.Sprintf("paso_images[%d]", i)
			fileHeaders := c.Request.MultipartForm.File[fileKey]
			if len(fileHeaders) > 0 {
				fileHeader := fileHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), newPaso.OrdenPaso, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err != nil {
					log.Printf("Error subiendo imagen paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo imagen de paso: " + err.Error()})
					return
				}
				pasoModel.URLImagen = url
			}

			// Upload video si proporcionado
			videoKey := fmt.Sprintf("paso_videos[%d]", i)
			videoHeaders := c.Request.MultipartForm.File[videoKey]
			if len(videoHeaders) > 0 {
				fileHeader := videoHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), newPaso.OrdenPaso, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err != nil {
					log.Printf("Error subiendo video paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo video de paso: " + err.Error()})
					return
				}
				pasoModel.URLVideo = url
			}

			// Guardar o crear
			if pasoModel.ID == 0 {
				if err := db.Debug().Create(&pasoModel).Error; err != nil {
					log.Printf("Error creando paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando paso: " + err.Error()})
					return
				}
			} else {
				if err := db.Debug().Save(&pasoModel).Error; err != nil {
					log.Printf("Error actualizando paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando paso: " + err.Error()})
					return
				}
			}
		}

		// Eliminar pasos que no están en la nueva lista
		newOrdens := make(map[int]bool)
		for _, np := range newPasos {
			newOrdens[np.OrdenPaso] = true
		}
		for orden, exist := range existingPasos {
			if !newOrdens[orden] {
				if err := db.Delete(&exist).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando paso obsoleto: " + err.Error()})
					return
				}
			}
		}
	}

	// Recargar material con relaciones
	db.Preload("Colaboradores").Preload("Galeria").Preload("Pasos").Preload("PropiedadesMecanicas").Preload("PropiedadesPerceptivas").Preload("PropiedadesEmocionales").Find(&material)

	c.JSON(http.StatusOK, material)
}
