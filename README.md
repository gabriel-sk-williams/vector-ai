# vector-ai

A fully-featured backend in production at caldera.sh. Suite includes user authorization, authentication and permissions, websockets, file upload, data vectorization, embedding, and query. Deployed with Gcloud.

Caldera is a AI platform-agnostic document query system, enabling users to upload personal documents and sync their own Google Drive stores. Each document is parsed, vectorized, and uploaded to a secure, private repository, allowing custom vector similarity searches on individual or groups of documents in tandem with calls to an LLM of their choosing.

### Directory structure

```
.
├── constants
│   └── constants.go           
├── drive                       // Implements Google Drive sync
│   └── controls.go          
│   └── drive_query.go             
│   └── drive_service.go           
├── middleware                  // JWT authentication
│   └── authenticate.go         
├── model                       // Struct and method definitions
│   └── context.go       
│   └── document.go      
│   └── response.go      
│   └── sql.go      
│   └── sync.go      
│   └── ws.go           
├── parse                       // Document parsing and upload to Qdrant
│   └── docx
│   └── epub 
│   └── pdf
│   └── txt 
├── postgres                    // Postgres module
│   └── ... (21 files)
├── qdrant                      // Vector database
│   └── controls.go
│   └── qdrant_grpc.go
│   └── qdrant_query.go
│   └── qdrant_vss.go
├── route
│   └── middleware.go           // User authorization
│   └── query_analysis.go       // Custom AI prompts and queries
│   └── query_vss.go            // Vector similarity search
│   └── route.go
│   └── session.go              // Individual websocket connections
│   └── sync_report.go          // Tracks Google drive synchronization
│   └── tracker.go              // Tracks websocket connections
│   └── upload_manual.go
│   └── upload_operations.go
│   └── upload_sync.go
├── util                        // Utility functions
│   └── util.go 
├── .gitignore
├── .application.go
├── config.json                 // Config file for local development
├── go.mod
├── go.sum
```


