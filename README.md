# 📚 HUB Innova - Material Library API

Backend robusto desarrollado para la gestión, catalogación y moderación de materiales didácticos y de laboratorio. Este proyecto implementa una arquitectura escalable con **Go**, gestión de roles jerárquicos (Lector/Colaborador/Admin), autenticación vía Google/Supabase y almacenamiento multimedia avanzado.

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-Framework-ff5a5f?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791?style=flat&logo=postgresql)
![GORM](https://img.shields.io/badge/ORM-GORM-red)

## 🚀 Tecnologías

* **Lenguaje:** Golang 1.23+
* **Web Framework:** Gin Gonic
* **ORM:** GORM (Driver Postgres)
* **Base de Datos & Auth:** Supabase (PostgreSQL + Google OAuth)
* **Storage:** Supabase Storage (Gestión de imágenes y videos para Pasos/Galerías)
* **Seguridad:** Validación de JWT (Supabase Auth) y Middleware RBAC.

## 🏗 Arquitectura

El proyecto sigue una estructura modular para facilitar el mantenimiento y la escalabilidad:

```text
├── api
│   ├── config       # Configuración de entorno y conexión DB string
│   ├── database     # Singleton de conexión GORM y cliente Supabase Storage
│   ├── handlers     # Controladores / Lógica de Negocio
│   │   ├── material # CRUD Materiales, Aprobaciones, Filtros
│   │   └── usuarios # Auth, Gestión de Usuarios, Notificaciones, Stats
│   ├── middleware   # Autenticación JWT y Control de Roles
│   └── models       # Structs y definición de tablas (con tipos JSONB custom)
│
├── go.mod           # Dependencias
└── main.go          # Punto de entrada, Router y Configuración CORS
```

✨ Funcionalidades Principales

🛡️ Gestión de Roles y Permisos (RBAC)
    Sistema de permisos granular para controlar el acceso a los recursos.

    Roles:

    👑 Administrador: Control total. Aprueba materiales, gestiona usuarios y ve métricas.

    👷 Colaborador: Puede crear y editar materiales. Requiere aprobación para publicar.

    👤 Lector: Acceso de solo lectura al catálogo público.

    Flujo de Roles: Los usuarios inician como "Lector" y pueden solicitar ser "Colaborador", lo cual notifica a los administradores.

🧪 Gestión de Materiales (Core)

    Modelo de datos complejo que soporta propiedades técnicas detalladas usando JSONB en PostgreSQL.

    Propiedades Dinámicas: Almacenamiento flexible de Composición, Herramientas, Propiedades Mecánicas, Perceptivas y Emocionales.

    Multimedia: Soporte para Galerías de imágenes y Pasos (Instrucciones) con fotos y videos, subidos directamente a Supabase Storage.

    Colaboración: Múltiples usuarios pueden aparecer como autores de un mismo material.

✅ Flujo de Moderación

    Sistema de Control de Calidad para los contenidos subidos.

    Estado Pendiente: Todo material nuevo o editado pasa a estado "Pendiente" (estado: false).

    Aprobación/Rechazo: Los administradores pueden aprobar (hacer público) o rechazar materiales (con motivo del rechazo).

    Notificaciones: Sistema de alertas internas para avisar sobre cambios de estado, solicitudes de rol o nuevos materiales.

👤 Gestión de Usuarios

    Auth Híbrida: Registro y Login manejado vía Token de Supabase (Google OAuth), sincronizado con la base de datos local.

    Administración: CRUD completo de usuarios con soporte para Soft Delete y Hard Delete (con validaciones de integridad referencial).

    Dashboard: Estadísticas rápidas sobre materiales y usuarios activos.

🛠️ Instalación y Configuración

    1. Clonar el repositorio: git clone 

        https://github.com/sllach/TT-SEM-2-back
        cd TT-SEM-2-BACK
    
    2. Configurar Variables de Entorno: Crea un archivo .env en la raíz con las siguientes credenciales:

        DB_HOST=""
        DB_PORT=""
        DB_USER=""
        DB_PASSWORD=""
        DB_NAME=""
        POOL_MODE=""
        PORT=""
        SUPABASE_SERVICE_KEY=""
        SUPABASE_PROJECT=""
        SUPABASE_URL=""
        SUPABASE_JWT_SECRET=""
        SUPABASE_ANON_KEY=""
    
    3. Instalar Dependencias

        go mod tidy

    4. Ejecutar Servidor

        go run main.go

## 📡 Endpoints API

### 🩺 Health & System

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/health` | Verificar estado del servidor y conexión a BD | 🟢 Público |

### 🔐 Auth & Perfil

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `POST` | `/auth/register` | Registro o Login con Token de Supabase | 🟢 Público |
| `GET` | `/me` | Obtener mi perfil y estadísticas | 🔵 Autenticado |
| `POST` | `/users/request-role` | Solicitar ascenso a Colaborador | 🔵 Lector |

### 🧪 Materiales (Público & Lectura)

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/materials` | Listar materiales aprobados (públicos) | 🟢 Público |
| `GET` | `/materials/:id` | Ver detalle completo de un material | 🟢 Público |
| `GET` | `/materials/:id/derived` | Ver materiales derivados de este ID | 🟢 Público |
| `GET` | `/materials/filters` | Obtener listas de herramientas y composiciones | 🟢 Público |
| `GET` | `/materials-summary` | Listado ligero para tarjetas (cards) | 🟢 Público |
| `GET` | `/users/:google_id/public` | Ver perfil público y materiales de un usuario | 🟢 Público |

### 📝 Gestión de Materiales (Creación)

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `POST` | `/materials` | Crear nuevo material (Multipart Form) | 🟠 Colaborador / Admin |
| `PUT` | `/materials/:id` | Editar material existente | 🟠 Colaborador / Admin |
| `DELETE` | `/materials/:id` | Eliminar material | 🔴 Admin |

### 👮 Moderación & Administración

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/materials/pending` | Ver materiales pendientes de revisión | 🔴 Admin |
| `POST` | `/materials/:id/approve` | Aprobar material (hacer público) | 🔴 Admin |
| `POST` | `/materials/:id/reject` | Rechazar material (con motivo) | 🔴 Admin |
| `GET` | `/users` | Listar todos los usuarios | 🔴 Admin |
| `GET` | `/users/:google_id` | Ver detalle de un usuario | 🔴 Admin |
| `PUT` | `/users/:google_id` | Editar usuario (Roles/Datos) | 🔴 Admin |
| `DELETE` | `/users/:google_id` | Soft Delete de usuario | 🔴 Admin |
| `DELETE` | `/users/:google_id/hard` | Hard Delete (Permanente) | 🔴 Admin |
| `GET` | `/users/stats` | KPIs del Dashboard | 🔴 Admin |

### 🔔 Notificaciones

| Método | Endpoint | Descripción | Nivel de Acceso |
| --- | --- | --- | --- |
| `GET` | `/notifications` | Listar mis notificaciones | 🟠 Colaborador / Admin |
| `PATCH` | `/notifications/:id/read` | Marcar notificación como leída | 🟠 Colaborador / Admin |
