package material

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateMaterial maneja la creación de un material
func CreateMaterial(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
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
	creadorIDStr := c.PostForm("creador_id")
	propMecanicasStr := c.PostForm("prop_mecanicas")
	propPerceptivasStr := c.PostForm("prop_perceptivas")
	propEmocionalesStr := c.PostForm("prop_emocionales")
	colaboradoresStr := c.PostForm("colaboradores")
	pasosStr := c.PostForm("pasos")
	galeriaCaptionsStr := c.PostForm("galeria_captions")

	// Validaciones básicas
	if nombre == "" || creadorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campos requeridos: nombre y creador_id"})
		return
	}

	creadorID, err := strconv.ParseUint(creadorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creador_id inválido"})
		return
	}

	// Parsear arrays desde strings JSON
	var herramientas models.StringArray
	if herramientasStr != "" {
		if err := json.Unmarshal([]byte(herramientasStr), &herramientas); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "herramientas inválido (debe ser JSON array)"})
			return
		}
	}
	var composicion models.StringArray
	if composicionStr != "" {
		if err := json.Unmarshal([]byte(composicionStr), &composicion); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "composicion inválido (debe ser JSON array)"})
			return
		}
	}

	var derivadoDe uuid.UUID
	if derivadoDeStr != "" {
		derivadoDe, err = uuid.Parse(derivadoDeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "derivado_de inválido (UUID)"})
			return
		}
	}

	// Crear el material
	material := models.Material{
		ID:           uuid.New(),
		Nombre:       nombre,
		Descripcion:  descripcion,
		Herramientas: herramientas,
		Composicion:  composicion,
		DerivadoDe:   derivadoDe,
		CreadorID:    uint(creadorID),
	}

	if err := db.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando material: " + err.Error()})
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
		if err := db.Debug().Create(&propMecanicas).Error; err != nil {
			log.Printf("Error creando prop_mecanicas: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando propiedades mecánicas: " + err.Error()})
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
		if err := db.Debug().Create(&propPerceptivas).Error; err != nil {
			log.Printf("Error creando prop_perceptivas: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando propiedades perceptivas: " + err.Error()})
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
		if err := db.Debug().Create(&propEmocionales).Error; err != nil {
			log.Printf("Error creando prop_emocionales: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando propiedades emocionales: " + err.Error()})
			return
		}
	}

	// Colaboradores
	if colaboradoresStr != "" {
		var colaboradores []uint
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

	// Parsear captions para galería (JSON array)
	var galeriaCaptions []string
	if galeriaCaptionsStr != "" {
		if err := json.Unmarshal([]byte(galeriaCaptionsStr), &galeriaCaptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "galeria_captions inválido (JSON array)"})
			return
		}
	}

	// Manejar uploads de galería
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
		if err := db.Debug().Create(&galeria).Error; err != nil {
			log.Printf("Error creando galeria: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando galería: " + err.Error()})
			return
		}
	}

	// Manejar pasos
	if pasosStr != "" {
		var pasos []struct {
			OrdenPaso   int    `json:"orden_paso"`
			Descripcion string `json:"descripcion"`
		}
		if err := json.Unmarshal([]byte(pasosStr), &pasos); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pasos inválido (JSON array)"})
			return
		}

		for i, paso := range pasos {
			pasoModel := models.PasoMaterial{
				MaterialID:  material.ID,
				OrdenPaso:   paso.OrdenPaso,
				Descripcion: paso.Descripcion,
			}

			// Upload imagen
			fileKey := fmt.Sprintf("paso_images[%d]", i)
			fileHeaders := c.Request.MultipartForm.File[fileKey]
			if len(fileHeaders) > 0 {
				fileHeader := fileHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err != nil {
					log.Printf("Error subiendo imagen paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo imagen de paso: " + err.Error()})
					return
				}
				pasoModel.URLImagen = url
			}

			// Upload video
			videoKey := fmt.Sprintf("paso_videos[%d]", i)
			videoHeaders := c.Request.MultipartForm.File[videoKey]
			if len(videoHeaders) > 0 {
				fileHeader := videoHeaders[0]
				safeFilename := strings.ReplaceAll(fileHeader.Filename, " ", "_")
				filePath := fmt.Sprintf("materials/%s/pasos/%d/%s", material.ID.String(), i, safeFilename)
				url, err := database.SubirAStorageSupabase(fileHeader, "pasos-bucket", filePath)
				if err != nil {
					log.Printf("Error subiendo video paso %d: %v", i, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo video de paso: " + err.Error()})
					return
				}
				pasoModel.URLVideo = url
			}

			if err := db.Debug().Create(&pasoModel).Error; err != nil {
				log.Printf("Error creando paso %d: %v", i, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando paso: " + err.Error()})
				return
			}
		}
	}

	// Recargar material con relaciones
	db.Preload("Colaboradores").Preload("Galeria").Preload("Pasos").Preload("PropiedadesMecanicas").Preload("PropiedadesPerceptivas").Preload("PropiedadesEmocionales").Find(&material)

	c.JSON(http.StatusCreated, material)
}
