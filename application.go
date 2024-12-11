package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"vector-ai/middleware"
	"vector-ai/model"
	"vector-ai/postgres"
	"vector-ai/qdrant"
	"vector-ai/route"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	pb "github.com/qdrant/go-client/qdrant"
	"github.com/unidoc/unipdf/v3/common/license"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/cors"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

func main() {
	fmt.Println("Goyave server active 1.4")

	_, err := os.Stat("dev.txt")
	runMode := ""

	if err == nil {
		fmt.Println("dev mode")
		godotenv.Load(".env")
		runMode = "dev"
	} else {
		fmt.Println("production mode")
		runMode = "production"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_URL"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	server := os.Getenv("QDRANT_DB")
	apiKey := os.Getenv("CLERK_API_KEY")

	clClient, err := clerk.NewClient(apiKey)

	if err != nil {
		fmt.Println(err)
	}

	var (
		addr = flag.String("addr", server, "the address to connect to")
	)

	qdDriver, conn := dialQdrant(addr)
	pgDriver := dialPostgres(connStr, runMode)
	migratePostgres(connStr, runMode)

	qdClient := qdrant.Qdr{
		Driver:     qdDriver,
		Connection: conn,
	}

	pgClient := postgres.Pgx{
		Driver: pgDriver,
	}

	// ws session tracker
	jt := route.NewTracker()

	handler := route.Handler{
		QD: qdClient,
		PG: pgClient,
		TR: jt,
		CL: clClient,
	}

	jt.SetHandler(handler)

	err = license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		panic(err)
	}

	// start registration route
	if err := goyave.Start(func(router *goyave.Router) {

		// defaults
		router.CORS(cors.Default())
		// router.Middleware(middleware.Authentication)
		// router.GlobalMiddleware(glog.CombinedLogMiddleware())

		go jt.Run()

		// websocket connection
		router.Get("/status/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}", jt.Serve) // auth via ws

		// rest api
		router.Get("/", handler.Greeting)
		router.Get("/health", handler.Health)
		router.Get("/status", handler.GetStatus)
		// router.Get("/qdrant/{qId}", handler.GetQdrantCollection)

		inviteRouter := router.Group()
		inviteRouter.Middleware(middleware.Authentication)
		inviteRouter.Get("/invite/{token}", handler.GetInvite) // happens first

		onboardRouter := router.Group()
		onboardRouter.Middleware(middleware.Authentication, handler.Invite)
		onboardRouter.Post("/association/user/{userId}/org/{orgId}", handler.CreateUserOrgAssociation) // happens second

		adminRouter := router.Group()
		adminRouter.Middleware(middleware.Authentication, handler.Admin)
		adminRouter.Get("/admin/org", handler.ListOrgs)
		adminRouter.Get("/admin/invite", handler.ListInvites)
		adminRouter.Delete("/admin/invite/{inviteId}", handler.DeleteInvite)
		adminRouter.Delete("/admin/invite/clear", handler.ClearExpiredInvites)

		userRouter := router.Group()
		userRouter.Middleware(middleware.Authentication, handler.Authorization) // handler.Permissions
		userRouter.Get("/user/{userId}/org", handler.ListOrgsByUserId)
		userRouter.Get("/user/{userId}/admin", handler.GetAdmin)
		userRouter.Post("/user/{userId}", handler.CreateUser)
		userRouter.Delete("/user/{userId}/org/{orgId}", handler.DeleteUserOrgAssociation) // needs permissions

		userRouter.Get("/user/{userId}/config", handler.ListUserConfigs)
		userRouter.Get("/user/{userId}/config/{configId}", handler.GetUserConfig)
		userRouter.Put("/user/{userId}/config/{configId}", handler.UpdateUserConfig)

		orgRouter := router.Group()
		orgRouter.Middleware(middleware.Authentication, handler.Authorization)
		orgRouter.Post("/org", handler.CreateOrg)
		orgRouter.Get("/org/{orgId}", handler.GetOrg)
		orgRouter.Put("/org/{orgId}", handler.UpdateOrg)
		orgRouter.Delete("/org/{orgId}", handler.DeleteOrg)
		orgRouter.Get("/org/{orgId}/users", handler.ListUsersByOrgId)
		orgRouter.Get("/org/{orgId}/role", handler.ListOrgRoleAssignmentsByOrgId)
		orgRouter.Get("/org/{orgId}/invite", handler.ListInvitesByOrgId)
		orgRouter.Post("/org/{orgId}/invite", handler.CreateInvite)
		orgRouter.Post("/org/{orgId}/subscription", handler.CreateOrgStripeSubscriptionAssociation)
		orgRouter.Get("/org/{orgId}/subscription", handler.GetOrgStripeSubscriptionAssociation)
		orgRouter.Delete("/org/{orgId}/subscription", handler.DeleteOrgStripeSubscriptionAssociation)

		orgRouter.Get("/org/{orgId}/config", handler.ListOrgConfigs)
		orgRouter.Get("/org/{orgId}/config/{configId}", handler.GetOrgConfig)
		orgRouter.Put("/org/{orgId}/config/{configId}", handler.UpdateOrgConfig)

		orgRouter.Get("/org/{orgId}/user/{userId}/role", handler.ListOrgRoleAssignmentsByUserId)
		orgRouter.Post("/org/{orgId}/user/{userId}/role", handler.CreateOrgRoleAssignment).Validate(model.OrgRoleAssignmentProps)
		// orgRouter.Delete("/org/{orgId}/role/{roleId}", handler.DeleteOrgRoleAssignment)
		// orgRouter.Put("/org/{orgId}/role/{roleId}/update", handler.UpdateOrgRoleAssignment)

		workRouter := router.Group()
		workRouter.Middleware(middleware.Authentication, handler.Authorization)
		workRouter.Get("/org/{orgId}/workspace", handler.ListWorkspaces)
		workRouter.Post("/org/{orgId}/workspace", handler.CreateWorkspace).Validate(model.WorkspaceCreateProps)
		workRouter.Get("/org/{orgId}/workspace/{workspaceId}", handler.GetWorkspace) // GetPostgresWorkspace
		workRouter.Delete("/org/{orgId}/workspace/{workspaceId}", handler.DeleteWorkspace)
		workRouter.Put("/org/{orgId}/workspace/{workspaceId}/clear", handler.ClearWorkspace)
		workRouter.Put("/org/{orgId}/workspace/{workspaceId}/update", handler.UpdateWorkspace)

		workRouter.Get("/org/{orgId}/workspace/{workspaceId}/config", handler.ListWorkspaceConfigs)
		workRouter.Get("/org/{orgId}/workspace/{workspaceId}/config/{propertyName}", handler.GetWorkspaceConfig)
		workRouter.Put("/org/{orgId}/workspace/{workspaceId}/config/{propertyName}", handler.UpdateWorkspaceConfig).Validate(model.WorkspaceConfigProps)

		convRouter := router.Group()
		convRouter.Middleware(middleware.Authentication, handler.Authorization)
		convRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation", handler.ListConversations)
		convRouter.Post("/org/{orgId}/workspace/{workspaceId}/conversation", handler.CreateConversation)
		convRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}", handler.GetConversation)
		convRouter.Delete("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}", handler.DeleteConversation)
		convRouter.Put("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/clear", handler.ClearConversation)
		convRouter.Put("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/update", handler.UpdateConversation)

		msgRouter := router.Group()
		msgRouter.Middleware(middleware.Authentication, handler.Authorization)
		msgRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message", handler.ListMessages)
		msgRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message/{messageId}", handler.GetMessage)
		msgRouter.Delete("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message/{messageId}", handler.DeleteMessage)

		docRouter := router.Group()
		docRouter.Middleware(middleware.Authentication, handler.Authorization)
		docRouter.Get("/org/{orgId}/workspace/{workspaceId}/document", handler.ListDocuments)
		docRouter.Get("/org/{orgId}/workspace/{workspaceId}/document/{documentId}", handler.GetDocument)
		docRouter.Get("/org/{orgId}/workspace/{workspaceId}/document/{documentId}/chunk", handler.ListChunks)
		docRouter.Delete("/org/{orgId}/workspace/{workspaceId}/document/{documentId}", handler.DeleteDocument)

		ctxRouter := router.Group()
		ctxRouter.Middleware(middleware.Authentication, handler.Authorization)
		ctxRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/document/{documentId}/context", handler.ListContextsByDocument)
		ctxRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message/{messageId}/context", handler.ListContextsByMessage)
		ctxRouter.Get("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message/{messageId}/context/{contextId}", handler.GetContext)
		ctxRouter.Delete("/org/{orgId}/workspace/{workspaceId}/conversation/{conversationId}/message/{messageId}/context/{contextId}", handler.DeleteContext)

		tmplRouter := router.Group()
		tmplRouter.Middleware(middleware.Authentication, handler.Authorization)
		tmplRouter.Get("/org/{orgId}/workspace/{workspaceId}/template", handler.ListTemplates)
		tmplRouter.Post("/org/{orgId}/workspace/{workspaceId}/template", handler.CreateTemplate)
		tmplRouter.Get("/org/{orgId}/workspace/{workspaceId}/template/{templateId}", handler.GetTemplate)
		tmplRouter.Put("/org/{orgId}/workspace/{workspaceId}/template/{templateId}", handler.UpdateTemplate)
		tmplRouter.Delete("/org/{orgId}/workspace/{workspaceId}/template/{templateId}", handler.DeleteTemplate)

		syncRouter := router.Group()
		syncRouter.Middleware(middleware.Authentication, handler.Authorization)
		syncRouter.Get("/org/{orgId}/workspace/{workspaceId}/drive", handler.ListDriveFolderDocs).Validate(model.DriveListProps)
		syncRouter.Get("/org/{orgId}/workspace/{workspaceId}/shared-drive", handler.ListSharedDrives)
		syncRouter.Get("/org/{orgId}/workspace/{workspaceId}/sync-report", handler.GetSyncReport)
		syncRouter.Delete("/org/{orgId}/workspace/{workspaceId}/reverse-sync", handler.ClearWorkspace)
		syncRouter.Delete("/org/{orgId}/workspace/{workspaceId}/drive/{driveSyncId}", handler.DeleteDriveFolderSync)
		syncRouter.Get("/org/{orgId}/workspace/{workspaceId}/sync-tree", handler.GetSyncTree)

		tagRouter := router.Group()
		tagRouter.Middleware(middleware.Authentication, handler.Authorization)
		tagRouter.Get("/org/{orgId}/tag", handler.ListTags)
		tagRouter.Post("/org/{orgId}/tag", handler.CreateTag)
		tagRouter.Get("/org/{orgId}/tag/{tagId}", handler.GetTag)
		tagRouter.Put("/org/{orgId}/tag/{tagId}", handler.UpdateTag)
		tagRouter.Delete("/org/{orgId}/tag/{tagId}", handler.DeleteTag)

		tagRouter.Get("/org/{orgId}/workspace/{workspaceId}/tag", handler.ListDocumentTagAssociations)
		tagRouter.Get("/org/{orgId}/workspace/{workspaceId}/tag/{tagId}/document", handler.ListDocumentsByTagId)
		tagRouter.Get("/org/{orgId}/workspace/{workspaceId}/document/{documentId}/tag", handler.ListTagsByDocumentId)
		tagRouter.Post("/org/{orgId}/workspace/{workspaceId}/document/{documentId}/tag/{tagId}", handler.TagDocument)     // create Association
		tagRouter.Delete("/org/{orgId}/workspace/{workspaceId}/document/{documentId}/tag/{tagId}", handler.UntagDocument) // delete Association
		tagRouter.Post("/org/{orgId}/workspace/{workspaceId}/document/{documentId}/tag/{tagId}", handler.GetDocumentTagAssociation)

		// TODO
		// bulk tag endpoint
		// bulk untag endpoint

	}); err != nil {
		os.Exit(err.(*goyave.Error).ExitCode)
	}
}

func dialPostgres(connStr string, runMode string) *pgxpool.Pool {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		fmt.Println(err)
	}

	if runMode == "production" {
		// Create a new dialer with any options
		d, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			fmt.Println(err)
		}

		// Tell the driver to use the Cloud SQL Go Connector to create connections
		config.ConnConfig.DialFunc = func(ctx context.Context, _ string, instance string) (net.Conn, error) {
			return d.Dial(ctx, os.Getenv("POSTGRES_CONN_NAME"))
		}
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v\n", err))
	}
	return db
}

func dialQdrant(addr *string) (pb.CollectionsClient, *grpc.ClientConn) {

	flag.Parse()
	// Set up a connection to the server.
	config := &tls.Config{}
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	collections_client := pb.NewCollectionsClient(conn)

	return collections_client, conn
}

func migratePostgres(connStr string, runMode string) {
	if runMode == "dev" {
		sql, err := goose.OpenDBWithDriver("pgx", connStr)
		if err != nil {
			log.Fatalf(err.Error())
		}

		println("Migrating")

		err = goose.Up(sql, "./db/migrations")
		if err != nil {
			log.Fatalf(err.Error())
		}
	} else {
		println("Not in Dev environment, skipping DB migrations")
	}
}
