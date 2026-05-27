# go-upload-files

Backend em Go para armazenamento de arquivos com estrutura hierárquica de pastas, inspirado no Google Drive. Permite que usuários autenticados façam upload de arquivos para o AWS S3, organizem o conteúdo em pastas e realizem operações completas de gerenciamento.

---

## Visão Geral

O projeto expõe uma API HTTP RESTful que cobre autenticação de usuários, gerenciamento de pastas e arquivos com suporte a upload multipart/resumable. Os binários são armazenados no AWS S3 e os metadados persistidos em PostgreSQL.

**Stack:**
- **Linguagem:** Go (stand libary)
- **Banco de dados:** PostgreSQL via GORM
- **Storage:** AWS S3 (SDK v2)
- **Autenticação:** JWT via cookie HttpOnly
- **Infraestrutura:** Terraform (provisionamento do bucket S3)

---

## Fluxo de Arquivos

![Fluxo de upload e download](assets/Fluxo%20go-upload-files.jpg)

### 1. Upload de Arquivo

O usuário inicia o processo chamando `POST /uploads/init`, que registra a sessão de upload. A API então envia o arquivo em partes diretamente para o **AWS S3** usando o protocolo multipart — ideal para arquivos grandes, pois cada parte é enviada de forma independente e pode ser retomada em caso de falha. Após a conclusão (`POST /uploads/{id}/complete`), os metadados do arquivo (nome, tamanho, tipo, localização no S3) são persistidos no **PostgreSQL**.

### 2. Download de Arquivo

O usuário solicita o download via `GET /files/{id}/download`. A API consulta o PostgreSQL para validar que o arquivo existe e pertence ao usuário, depois gera uma **URL assinada temporária** diretamente no AWS S3. Essa URL é retornada ao usuário, que faz o download direto do S3 sem passar pelo servidor — reduzindo latência e consumo de banda da API.

---

## Arquitetura

```
Client
  → HTTP API (net/http)
  → Middleware (CORS, Auth, Error)
  → Handlers
  → Services
  → Repositories
  → PostgreSQL + AWS S3
```

### Estrutura de diretórios

```
app/
├── cmd/api-server/       # Entrypoint da aplicação
├── api/routes/           # Registro de rotas
├── internal/
│   ├── database/         # Conexão com PostgreSQL (GORM)
│   ├── dto/              # Objetos de transferência de dados
│   ├── handlers/         # Handlers HTTP
│   ├── middleware/       # CORS, Auth JWT, Error handling
│   ├── models/           # Modelos de domínio (User, Folder, File)
│   ├── repositories/     # Acesso ao banco de dados
│   ├── services/         # Regras de negócio
│   └── storage/aws/      # Integração com S3
└── go.mod
infra/
└── terraform/            # Provisionamento do bucket S3
```

---

## Funcionalidades

### Autenticação
- Cadastro de usuário com nome, email e senha (hash bcrypt)
- Login com emissão de JWT via cookie `HttpOnly` + `SameSite=Strict`
- Logout com limpeza do cookie no backend
- Todas as rotas de arquivos e pastas exigem autenticação

### Pastas
- Criar pasta raiz ou subpasta dentro de outra pasta
- Listar todas as pastas do usuário
- Buscar pasta por ID
- Listar subpastas imediatas de uma pasta
- Obter o caminho completo (breadcrumb) de uma pasta
- Listar conteúdo imediato de uma pasta (subpastas + arquivos)
- Renomear pasta
- Mover pasta para outro pai (com validação de ciclo)
- Excluir pasta recursivamente (subpastas e arquivos)

### Arquivos
- Upload multipart/resumable: iniciar, enviar partes, concluir e abortar
- Listar arquivos do usuário com filtros (nome, status, período, pasta, paginação)
- Buscar arquivo por ID
- Obter URL assinada para download
- Listar arquivos de uma pasta específica
- Renomear arquivo
- Editar metadados básicos
- Mover arquivo entre pastas
- Exclusão lógica de um ou vários arquivos

---

## Modelo de Domínio

**User** — usuário autenticado; email único; senha nunca armazenada em texto puro.

**Folder** — pasta com suporte a hierarquia via `parent_id`; `parent_id = null` representa a raiz do usuário; exclusão lógica via `deleted_at`.

**File** — arquivo com metadados (nome, mime type, tamanho, storage key, URL, status); `folder_id = null` representa arquivo na raiz; exclusão lógica via `deleted_at`. Status possíveis: `UPLOADED`, `READY`, `FAILED`, `DELETED`.

**UploadSession** — sessão de upload multipart, rastreia o `upload_id` do S3 e as partes enviadas.

---

## API

### Auth

| Método | Rota             | Descrição           |
|--------|------------------|---------------------|
| POST   | /auth/register   | Cadastrar usuário   |
| POST   | /auth/login      | Autenticar usuário  |
| POST   | /auth/logout     | Encerrar sessão     |

### Arquivos

| Método | Rota                              | Descrição                        |
|--------|-----------------------------------|----------------------------------|
| POST   | /uploads/init                     | Iniciar upload multipart         |
| PUT    | /uploads/{uploadId}/parts/{part}  | Enviar parte do upload           |
| POST   | /uploads/{uploadId}/complete      | Concluir upload multipart        |
| DELETE | /uploads/{uploadId}               | Abortar upload                   |
| GET    | /files                            | Listar arquivos com filtros      |
| GET    | /files/{fileId}                   | Buscar arquivo por ID            |
| GET    | /files/{fileId}/download          | Obter URL de download            |
| GET    | /folders/{folderId}/files         | Listar arquivos de uma pasta     |
| PATCH  | /files/{fileId}/name              | Renomear arquivo                 |
| PATCH  | /files/{fileId}/metadata          | Editar metadados                 |
| PATCH  | /files/{fileId}/folder            | Mover arquivo                    |
| DELETE | /files                            | Excluir um ou vários arquivos    |

### Pastas

| Método | Rota                              | Descrição                        |
|--------|-----------------------------------|----------------------------------|
| POST   | /folders                          | Criar pasta raiz                 |
| POST   | /folders/{folderId}/children      | Criar subpasta                   |
| GET    | /folders                          | Listar pastas do usuário         |
| GET    | /folders/{folderId}               | Buscar pasta por ID              |
| GET    | /folders/{folderId}/children      | Listar subpastas imediatas       |
| GET    | /folders/{folderId}/path          | Obter caminho completo           |
| GET    | /folders/{folderId}/items         | Listar subpastas + arquivos      |
| PATCH  | /folders/{folderId}/name          | Renomear pasta                   |
| PATCH  | /folders/{folderId}/parent        | Mover pasta                      |
| DELETE | /folders/{folderId}               | Excluir pasta recursivamente     |

### Padrão de resposta

```json
// Sucesso
{ "data": {}, "message": "optional" }

// Erro
{ "error": { "code": "RESOURCE_NOT_FOUND", "message": "folder not found" } }
```

---

## Configuração

Crie o arquivo `app/.env` com as seguintes variáveis:

```env
PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=seu_usuario
DB_PASSWORD=sua_senha
DB_NAME=nome_do_banco

# AWS
AWS_REGION=us-east-2
AWS_ACCESS_KEY_ID=sua_access_key
AWS_SECRET_ACCESS_KEY=sua_secret_key
AWS_BUCKET_NAME=nome_do_bucket

# JWT
JWT_SECRET=seu_segredo_jwt
```

---

## Como executar

```bash
cd app
go run ./cmd/api-server
```

O servidor sobe na porta definida em `PORT` (padrão: `8080`). A rota `GET /health` pode ser usada para verificar se está no ar.

---

## Infraestrutura (Terraform)

O bucket S3 é provisionado via Terraform com versionamento habilitado, regra de lifecycle para abortar uploads multipart incompletos após 3 dias e configuração de CORS.

```bash
cd infra/terraform
terraform init
terraform apply
```

---

## Segurança

- Senhas hasheadas com bcrypt
- JWT transmitido exclusivamente via cookie `HttpOnly` + `SameSite=Strict`; `Secure` ativado em produção
- O `ownerId` é sempre derivado do token JWT, nunca aceito do cliente
- Cada usuário acessa apenas seus próprios recursos
- URLs de download assinadas com expiração limitada
