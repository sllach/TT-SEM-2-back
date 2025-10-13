package database

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SubirAStorageSupabase sube un archivo directamente al bucket de Supabase Storage
func SubirAStorageSupabase(fileHeader *multipart.FileHeader, bucketName, filePath string) (string, error) {
	supabaseProject := os.Getenv("SUPABASE_PROJECT")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	if supabaseProject == "" || supabaseServiceKey == "" {
		return "", fmt.Errorf("variables de entorno SUPABASE_PROJECT o SUPABASE_SERVICE_KEY no configuradas")
	}

	// 1. Abrir el archivo
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el archivo del form: %v", err)
	}
	defer file.Close()

	// 2. Leer el contenido del archivo
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error leyendo el archivo: %v", err)
	}

	// 3. Detectar MIME type
	mimeType := http.DetectContentType(fileBytes)

	if mimeType == "application/octet-stream" {
		ext := filepath.Ext(fileHeader.Filename)
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		case ".mp4":
			mimeType = "video/mp4"
		case ".mov":
			mimeType = "video/quicktime"
		}
	}

	log.Printf("Subiendo archivo: %s (MIME: %s, Size: %d bytes)", fileHeader.Filename, mimeType, len(fileBytes))

	// 4. Construir la URL
	supabaseURL := fmt.Sprintf("https://%s.supabase.co", supabaseProject)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, filePath)

	log.Printf("Upload URL: %s", uploadURL)

	// 5. Crear la petición POST con el archivo en el body
	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("error creando request a supabase: %v", err)
	}

	// Headers para Supabase Storage
	req.Header.Set("Authorization", "Bearer "+supabaseServiceKey)
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	req.Header.Set("x-upsert", "true")

	// 6. Enviar la petición
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error al enviar request a supabase: %v", err)
	}
	defer resp.Body.Close()

	// 7. Leer respuesta
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// 8. Validar respuesta
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Error en respuesta de Supabase: status %d, body: %s", resp.StatusCode, bodyStr)
		return "", fmt.Errorf("error al subir archivo (status %d): %s", resp.StatusCode, bodyStr)
	}

	log.Printf("Archivo subido exitosamente. Response: %s", bodyStr)

	// 9. Construir la URL pública del archivo
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, filePath)

	return publicURL, nil
}
