package material

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateMaterial maneja la actualización de un material con form-data (para imágenes/videos en Supabase Storage)
func UpdateMaterial(c *gin.Context) {
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

	var material models.Material
	if err := db.First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Parsear form-data (max 32MB para uploads)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parseando form-data: " + err.Error()})
		return
	}

	// Actualizar campos si se envían
	if nombre := c.PostForm("nombre"); nombre != "" {
		material.Nombre = nombre
	}

	if descripcion := c.PostForm("descripcion"); descripcion != "" {
		material.Descripcion = descripcion
	}

	if herramientasStr := c.PostForm("herramientas"); herramientasStr != "" {
		var herramientas models.StringArray
		if err := json.Unmarshal([]byte(herramientasStr), &herramientas); err == nil {
			material.Herramientas = herramientas
		}
	}

	if composicionStr := c.PostForm("composicion"); composicionStr != "" {
		var composicion models.StringArray
		if err := json.Unmarshal([]byte(composicionStr), &composicion); err == nil {
			material.Composicion = composicion
		}
	}

	if derivadoDeStr := c.PostForm("derivado_de"); derivadoDeStr != "" {
		derivadoDe, err := uuid.Parse(derivadoDeStr)
		if err == nil {
			material.DerivadoDe = derivadoDe
		}
	}

	if creadorIDStr := c.PostForm("creador_id"); creadorIDStr != "" {
		creadorID, err := strconv.ParseUint(creadorIDStr, 10, 32)
		if err == nil {
			material.CreadorID = uint(creadorID)
		}
	}

	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando material: " + err.Error()})
		return
	}

	// Actualizar propiedades mecánicas si se envía (sobrescribe si existe)
	propMecanicasStr := c.PostForm("prop_mecanicas")
	if propMecanicasStr != "" {
		var propMecanicas models.PropiedadesMecanicas
		if err := json.Unmarshal([]byte(propMecanicasStr), &propMecanicas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prop_mecanicas inválido (JSON object)"})
			return
		}
		propMecanicas.MaterialID = material.ID
		db.Where("material_id = ?", material.ID).Delete(&models.PropiedadesMecanicas{}) // Elimina anterior si existe
		if err := db.Create(&propMecanicas).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando propiedades mecánicas: " + err.Error()})
			return
		}
	}

	// Similar para prop_perceptivas y prop_emocionales (copia el bloque arriba, cambia nombres)

	// Colaboradores: Reemplaza todos si se envían
	colaboradoresStr := c.PostForm("colaboradores")
	if colaboradoresStr != "" {
		var colaboradores []uint
		if err := json.Unmarshal([]byte(colaboradoresStr), &colaboradores); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "colaboradores inválido (JSON array)"})
			return
		}
		db.Where("material_id = ?", material.ID).Delete(&models.ColaboradorMaterial{}) // Elimina anteriores
		for _, userID := range colaboradores {
			colaborador := models.ColaboradorMaterial{
				MaterialID: material.ID,
				UsuarioID:  userID,
			}
			if err := db.Create(&colaborador).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando colaborador: " + err.Error()})
				return
			}
		}
	}

	// Parsear captions para galería (JSON array)
	var galeriaCaptions []string
	galeriaCaptionsStr := c.PostForm("galeria_captions")
	if galeriaCaptionsStr != "" {
		if err := json.Unmarshal([]byte(galeriaCaptionsStr), &galeriaCaptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "galeria_captions inválido (JSON array)"})
			return
		}
	}

	// Agregar nuevas imágenes a galería (no elimina antiguas)
	files := c.Request.MultipartForm.File["galeria_images[]"]
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
		if err := db.Create(&galeria).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando galería: " + err.Error()})
			return
		}
	}

	// Actualizar pasos: Reemplaza todos si se envían
	pasosStr := c.PostForm("pasos")
	if pasosStr != "" {
		var pasos []struct {
			OrdenPaso   int    `json:"orden_paso"`
			Descripcion string `json:"descripcion"`
		}
		if err := json.Unmarshal([]byte(pasosStr), &pasos); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pasos inválido (JSON array)"})
			return
		}

		db.Where("material_id = ?", material.ID).Delete(&models.PasoMaterial{}) // Elimina anteriores

		for i, paso := range pasos {
			pasoModel := models.PasoMaterial{
				MaterialID:  material.ID,
				OrdenPaso:   paso.OrdenPaso,
				Descripcion: paso.Descripcion,
			}

			// Upload nueva imagen si se envía
			fileKey := fmt.Sprintf("paso_images[%d]", i)
			fileHeaders := c.Request.MultipartForm.File[fileKey]
			if len(fileHeaders) > 0 {
				fileHeader := fileHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err == nil {
					pasoModel.URLImagen = url
				}
			}

			// Upload nuevo video si se envía
			videoKey := fmt.Sprintf("paso_videos[%d]", i)
			videoHeaders := c.Request.MultipartForm.File[videoKey]
			if len(videoHeaders) > 0 {
				fileHeader := videoHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err == nil {
					pasoModel.URLVideo = url
				}
			}

			if err := db.Create(&pasoModel).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando paso: " + err.Error()})
				return
			}
		}
	}

	// Recargar material con relaciones
	db.Preload("Colaboradores").Preload("Galeria").Preload("Pasos").Preload("PropiedadesMecanicas").Preload("PropiedadesPerceptivas").Preload("PropiedadesEmocionales").Find(&material)

	// Responder
	c.JSON(http.StatusOK, material)
}
