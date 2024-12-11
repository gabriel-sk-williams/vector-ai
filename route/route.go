package route

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"vector-ai/drive"
	"vector-ai/model"
	pgx "vector-ai/postgres"
	"vector-ai/qdrant"
	"vector-ai/util"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/go-errors/errors"

	"github.com/tmc/langchaingo/embeddings"
	"google.golang.org/api/googleapi"
	"goyave.dev/goyave/v4"
)

type Handler struct {
	QD  qdrant.Controls
	PG  pgx.Controls
	DRV drive.Controls
	TR  *Tracker
	CL  clerk.Client
	EM  *embeddings.EmbedderImpl
}

// curl http://localhost:5000
func (h Handler) Greeting(res *goyave.Response, req *goyave.Request) {
	res.String(http.StatusOK, "Hey Sonic!")
}

func (h Handler) Health(res *goyave.Response, req *goyave.Request) {
	res.Status(http.StatusOK)
}

// curl http://localhost:5000/status returns 1.4.1
func (h Handler) GetStatus(res *goyave.Response, req *goyave.Request) {
	status, err := h.QD.GetStatus()

	if err == nil {
		res.String(http.StatusOK, status)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// curl http://localhost:5000/qdrant/2ce911d8-db4c-451c-8aa4-144794361dfb
func (h Handler) GetQdrantCollection(res *goyave.Response, req *goyave.Request) {
	result, err := h.QD.GetCollection(req.Params["qId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Invites
//

func (h Handler) ListInvites(res *goyave.Response, req *goyave.Request) {
	invites, err := h.PG.ListInvites()

	if err == nil {
		res.JSON(http.StatusOK, invites)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetInvite(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetInvite(req.Params["token"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteInvite(res *goyave.Response, req *goyave.Request) {
	inviteId := req.Params["inviteId"]
	err := h.PG.DeleteMessage(inviteId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted invite %s", inviteId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ClearExpiredInvites(res *goyave.Response, req *goyave.Request) {

	invites, err := h.PG.ListInvites()
	check(err)

	var cleared int
	for _, invite := range invites {
		now := time.Now()
		expiry := invite.Expiration

		if now.After(expiry) {
			err = h.PG.DeleteInvite(invite.ID)
			if err == nil {
				cleared += 1
			}
		}
	}

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted %d invites ", cleared)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// User Routes
//

// curl http://localhost:5000/user/user_2Tm5PfVIX9uedCfsVe5zsJDniJw/admin
func (h Handler) GetAdmin(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	orgs, err := h.PG.ListOrgsByUserId(userId)

	isAdmin := util.IsAdmin(orgs)

	if err == nil {
		res.JSON(http.StatusOK, model.AdminResponse{IsAdmin: isAdmin})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateUser(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	firstName := req.String("firstName")
	lastName := req.String("lastName")
	email := req.String("email")

	fmt.Println("new user:", userId, firstName, lastName, email)

	user, err := h.PG.CreateUser(userId, firstName, lastName, email)
	check(err)

	if err == nil {
		res.JSON(http.StatusCreated, user)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// TODO: check permissions
func (h Handler) DeleteUserOrgAssociation(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	orgId := req.Params["orgId"]

	err := h.PG.DeleteUserOrgAssociation(orgId, userId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted user %s", userId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateUserOrgAssociation(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	orgId := req.Params["orgId"]

	token := req.String("token")
	invite, err := h.PG.GetInvite(token)
	check(err)

	fmt.Println(invite)

	result, err := h.PG.CreateUserOrgAssociation(userId, orgId)

	// consume invite
	if err == nil {
		err = h.PG.DeleteInvite(invite.ID)
	}

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusNotModified) // 304
		res.Error(err)
	}
}

//
// Org Configs
//

func (h Handler) ListOrgConfigs(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	configs, err := h.PG.ListOrgConfigs(orgId)

	if err == nil {
		res.JSON(http.StatusCreated, configs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetOrgConfig(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	result, err := h.PG.GetOrgConfig(orgId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateOrgConfig(res *goyave.Response, req *goyave.Request) {
	configId := req.Params["configId"]
	value := req.String("value")

	result, err := h.PG.UpdateOrgConfig(configId, value)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Org Role Assignments
//

// orgRouter.Get("/org/{orgId}/role", handler.ListOrgRoleAssignments)
func (h Handler) ListOrgRoleAssignmentsByOrgId(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	configs, err := h.PG.ListOrgRoleAssignments(orgId)

	if err == nil {
		res.JSON(http.StatusCreated, configs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListOrgRoleAssignmentsByUserId(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	userId := req.Params["userId"]
	configs, err := h.PG.ListOrgRoleAssignmentsByUserId(orgId, userId)

	if err == nil {
		res.JSON(http.StatusCreated, configs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// mass update
func (h Handler) CreateOrgRoleAssignment(res *goyave.Response, req *goyave.Request) {

	orgId := req.Params["orgId"]
	role := req.String("role")
	userIds := req.Data["templates"].([]string)

	roleAssignments := []model.OrgRoleAssignment{}
	for _, userId := range userIds {
		role, err := h.PG.CreateOrgRoleAssignment(orgId, userId, role)
		if err == nil {
			roleAssignments = append(roleAssignments, role)
		} else {
			res.Status(http.StatusInternalServerError)
			res.Error(err)
			return
		}
	}

	res.JSON(http.StatusCreated, roleAssignments)
}

func (h Handler) GetOrgRoleAssignmentById(res *goyave.Response, req *goyave.Request) {
	roleId := req.Params["roleId"]

	result, err := h.PG.GetOrgRoleAssignmentById(roleId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetOrgRoleAssignmentByOrgUser(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	userId := req.Params["userId"]

	result, err := h.PG.GetOrgRoleAssignmentByOrgUser(orgId, userId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateOrgRoleAssignment(res *goyave.Response, req *goyave.Request) {
	roleId := req.Params["roleId"]
	role := req.String("role")

	result, err := h.PG.UpdateOrgRoleAssignment(roleId, role)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteOrgRoleAssignment(res *goyave.Response, req *goyave.Request) {
	roleId := req.Params["roleId"]
	err := h.PG.DeleteOrgRoleAssignment(roleId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted org role assigment %s", roleId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Workspace Configs
//

func (h Handler) ListWorkspaceConfigs(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	configs, err := h.PG.ListWorkspaceConfigs(workspaceId)

	if err == nil {
		res.JSON(http.StatusCreated, configs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetWorkspaceConfig(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	propertyName := req.Params["propertyName"]
	result, err := h.PG.GetWorkspaceConfig(workspaceId, propertyName)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateWorkspaceConfig(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	propertyName := req.Params["propertyName"]
	value := int64(req.Integer("value"))

	// value := int64(req.Integer("value"))

	result, err := h.PG.UpdateWorkspaceConfig(workspaceId, propertyName, value)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// UserConfigs
//

func (h Handler) ListUserConfigs(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	configs, err := h.PG.ListUserConfigs(userId)

	if err == nil {
		res.JSON(http.StatusCreated, configs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetUserConfig(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	result, err := h.PG.GetUserConfig(userId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateUserConfig(res *goyave.Response, req *goyave.Request) {
	configId := req.Params["configId"]
	value := req.String("value")

	result, err := h.PG.UpdateUserConfig(configId, value)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Orgs
//

// curl http://localhost:5000/org
func (h Handler) ListOrgs(res *goyave.Response, req *goyave.Request) {
	orgs, err := h.PG.ListOrgs()

	if err == nil {
		res.JSON(http.StatusOK, orgs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// curl http://localhost:5000/user/user_2Tm5PfVIX9uedCfsVe5zsJDniJw/orgs/List
func (h Handler) ListOrgsByUserId(res *goyave.Response, req *goyave.Request) {
	userId := req.Params["userId"]
	orgs, err := h.PG.ListOrgsByUserId(userId)

	if err == nil {
		res.JSON(http.StatusOK, orgs)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetOrg(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	result, err := h.PG.GetOrg(orgId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateOrg(res *goyave.Response, req *goyave.Request) {

	orgName := req.String("name")
	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject

	// check for admin privileges to make more than one org
	orgs, err := h.PG.ListOrgsByUserId(userId)
	isAdmin := util.IsAdmin(orgs)

	// FORBIDDEN for non-subscribers
	if !isAdmin && len(orgs) > 0 {
		res.Status(http.StatusPaymentRequired)
		res.Error(err)
		return
	}

	org, err := h.PG.CreateOrg(orgName)
	check(err)

	orgId := org.ID
	_, err = h.PG.CreateUserOrgAssociation(userId, orgId)

	// check for collection before making
	_, qErr := h.QD.GetCollection(orgId)

	if qErr != nil {
		h.QD.CreateCollection(orgId, uint64(1536))
	}

	if err == nil {
		res.JSON(http.StatusCreated, org)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateOrg(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	var org model.Org
	var err error

	if req.Has("level") {
		orgLevel := req.String("level")
		org, err = h.PG.UpdateOrgLevel(orgId, orgLevel)
	}

	if req.Has("name") {
		orgName := req.String("name")
		org, err = h.PG.UpdateOrgName(orgId, orgName)
	}

	if err == nil {
		res.JSON(http.StatusOK, org)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteOrg(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]

	// Get all workspaces belonging to org
	workspaces, err := h.PG.ListWorkspaces(orgId)
	check(err)

	// Delete all workspaces
	for _, workspace := range workspaces {
		err = h.PG.DeleteWorkspace(workspace.ID)
		check(err)
	}

	// Get all userOrgAssociations
	userOrgAssociations, err := h.PG.ListUserOrgAssociations(orgId)
	check(err)

	// Delete all userOrgAssocations
	for _, userOrgAssociation := range userOrgAssociations {
		err = h.PG.DeleteUserOrgAssociation(userOrgAssociation.OrgID, userOrgAssociation.UserID)
		check(err)
	}

	// Delete Org
	err = h.PG.DeleteOrg(orgId)
	check(err)

	// Delete qdrant collection
	err = h.QD.DeleteCollection(orgId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted org %s", orgId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListUsersByOrgId(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	users, err := h.PG.ListUsers(orgId)

	if err == nil {
		res.JSON(http.StatusOK, users)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateInvite(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	email := req.String("email")

	// Create new Org for invitee if necessary
	if len(orgId) == 0 { // admin has selected "New Org"
		org, err := h.PG.CreateOrg(email)
		check(err)
		orgId = org.ID
		fmt.Println("Created new Org for onboarding", orgId)
	}

	expiration := time.Now().Add(time.Hour * 168).Format(time.RFC3339)
	result, err := h.PG.CreateInvite(orgId, email, expiration)
	check(err)

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListInvitesByOrgId(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]

	invites, err := h.PG.ListInvitesByOrgId(orgId)

	if err == nil {
		res.JSON(http.StatusOK, invites)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateOrgStripeSubscriptionAssociation(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	stripeId := req.String("subscriptionId")

	subscription, err := h.PG.CreateOrgStripeSubscriptionAssociation(stripeId, orgId, true)

	if err == nil {
		res.JSON(http.StatusCreated, subscription)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetOrgStripeSubscriptionAssociation(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]

	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)

	if err == nil {
		res.JSON(http.StatusCreated, subscription)
	} else {
		res.Status(http.StatusNotFound)
		res.Error(err)
	}
}

func (h Handler) DeleteOrgStripeSubscriptionAssociation(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	err := h.PG.DeleteOrgStripeSubscriptionAssociation(orgId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted subscription for org: %s", orgId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Workspaces
//

// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace
func (h Handler) ListWorkspaces(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	results, err := h.PG.ListWorkspaces(orgId)

	// Get LastMessage for each workspace, add as metadata
	workspacesResponse := []model.WorkspaceResponse{}
	for i, workspace := range results {

		lastMessage, err := h.PG.GetLastMessage(workspace.ID)
		softCheck(err)

		documents, err := h.PG.ListDocuments(workspace.ID)
		check(err)

		cr := model.WorkspaceResponse{
			Workspace:         results[i],
			LastMessage:       lastMessage,
			NumberOfDocuments: int64(len(documents)),
			DocumentsSize:     util.ReduceFileSize(documents),
		}
		workspacesResponse = append(workspacesResponse, cr)
	}

	// sort workspaces by most recent LastMessage
	slices.SortFunc(workspacesResponse,
		func(a model.WorkspaceResponse, b model.WorkspaceResponse) int {
			return b.LastMessage.Timestamp.Compare(a.LastMessage.Timestamp)
		})

	if err == nil {
		res.JSON(http.StatusOK, workspacesResponse)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace/col_Conan
func (h Handler) GetWorkspace(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetWorkspace(req.Params["workspaceId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateWorkspace(res *goyave.Response, req *goyave.Request) {

	// Get user information
	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject
	authorName, err := h.PG.GetUserName(userId)
	check(err)

	// Add to workspaces table
	orgId := req.Params["orgId"]
	workspaceName := req.String("name")

	// check if user is subscribed and limit to 5 workspaces
	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
	if !subscription.Active {
		workspaces, _ := h.PG.ListWorkspaces(orgId)
		if len(workspaces) > 4 {
			res.Status(http.StatusPaymentRequired)
			res.Error(err)
			return
		}
	}

	workspace, err := h.PG.CreateWorkspace(orgId, workspaceName)
	check(err)

	_, err = h.PG.CreateWorkspaceConfig("5a77587c-cfac-4726-8e15-35648f3db95a", workspace.ID, "vssDocumentLimit", 40)
	check(err)

	_, err = h.PG.CreateWorkspaceConfig("9a6b5324-1bee-46c5-862d-2289299b89d1", workspace.ID, "vssChunkLimit", 2)
	check(err)

	templates := req.Data["templates"].([]string)
	timestamp := time.Now().Format(time.RFC3339)

	// iterate through templateId's and copy for current org
	for _, templateId := range templates {
		template, err := h.PG.GetTemplate(templateId)
		check(err)
		h.PG.CreateTemplate(orgId, workspace.ID, template.Name, template.Text, userId, authorName, timestamp)
	}

	if err == nil {
		res.JSON(http.StatusCreated, workspace)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateWorkspace(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	workspaceName := req.String("name")

	result, err := h.PG.UpdateWorkspace(workspaceId, workspaceName)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace
func (h Handler) ClearWorkspace(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]

	// Clear documents from postgres
	err := h.PG.ClearDocuments(workspaceId)

	if err != nil {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}

	// Clear points from qdrant
	pointsDeleted, err := h.QD.DeleteVectorsByWorkspaceId(orgId, workspaceId)
	message := fmt.Sprintf("Cleared workspace, deleting '%d' points", pointsDeleted)

	if err == nil {
		res.JSON(http.StatusOK, message)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteWorkspace(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]
	err := h.PG.DeleteWorkspace(workspaceId)
	if err != nil {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}

	_, err = h.QD.DeleteVectorsByWorkspaceId(orgId, workspaceId)
	check(err)

	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
	if subscription.Active {
		record, err := h.createUsageEvent(orgId)
		fmt.Println(record)
		check(err)
	}

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted workspace %s", workspaceId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Conversations
//

func (h Handler) ListConversations(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	results, err := h.PG.ListConversations(workspaceId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetConversation(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetConversation(req.Params["conversationId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateConversation(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	conversationName := req.String("name")
	result, err := h.PG.CreateConversation(workspaceId, conversationName)

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateConversation(res *goyave.Response, req *goyave.Request) {
	conversationId := req.Params["conversationId"]
	conversationName := req.String("name")

	result, err := h.PG.UpdateConversation(conversationId, conversationName)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ClearConversation(res *goyave.Response, req *goyave.Request) {
	conversationId := req.Params["conversationId"]

	err := h.PG.ClearMessages(conversationId)

	if err == nil {
		res.JSON(http.StatusOK, "Cleared conversation")
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteConversation(res *goyave.Response, req *goyave.Request) {
	conversationId := req.Params["conversationId"]

	err := h.PG.ClearMessages(conversationId)
	check(err)

	err = h.PG.DeleteConversation(conversationId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted conversation %s", conversationId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Messages
//

func (h Handler) ListMessages(res *goyave.Response, req *goyave.Request) {
	results, err := h.PG.ListMessages(req.Params["conversationId"])

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetMessage(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetMessage(req.Params["messageId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteMessage(res *goyave.Response, req *goyave.Request) {
	messageId := req.Params["messageId"]
	err := h.PG.DeleteMessage(messageId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted message %s", messageId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Documents
//

func (h Handler) ListDocuments(res *goyave.Response, req *goyave.Request) {
	results, err := h.PG.ListDocuments(req.Params["workspaceId"])

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetDocument(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetDocument(req.Params["documentId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteDocument(res *goyave.Response, req *goyave.Request) {

	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]
	documentId := req.Params["documentId"]

	// Delete document from Qdrant
	pointsDeleted, err := h.QD.DeleteVectorsByDocumentId(orgId, workspaceId, documentId)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: develop strategy to properly desync route
	err = h.PG.DeleteDriveDocumentSyncByDocumentId(documentId)
	check(err)

	// Delete document from Postgres
	err = h.PG.DeleteDocument(documentId)
	check(err)

	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
	if subscription.Active {
		record, err := h.createUsageEvent(orgId)
		fmt.Println(record)
		check(err)
	}

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted %s (%d points)", documentId, pointsDeleted)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Context
//

func (h Handler) ListContextsByDocument(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	documentId := req.Params["documentId"]
	results, err := h.PG.ListContextsByDocumentId(documentId) // []Context{id, documentId, messageId, pointId}
	check(err)

	ids := []string{}
	for _, context := range results {
		ids = append(ids, context.PointID)
	}

	points, err := h.QD.GetPointsByUuid(workspaceId, ids)
	check(err)

	contexts := []model.ConsumableContext{}
	for _, point := range points {
		pointId := point.Id.GetUuid()
		payload := point.GetPayload()
		documentId := payload["documentId"].GetStringValue()
		text := payload["chunk"].GetStringValue()

		cc := model.ConsumableContext{ID: pointId, WorkspaceID: workspaceId, DocumentID: documentId, Text: text}
		contexts = append(contexts, cc)
	}

	if err == nil {
		res.JSON(http.StatusOK, contexts)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListContextsByMessage(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	messageId := req.Params["messageId"]
	results, err := h.PG.ListContextsByMessageId(messageId) // []Context{id, documentId, messageId, pointId}
	check(err)

	ids := []string{}
	for _, context := range results {
		ids = append(ids, context.PointID)
	}

	points, err := h.QD.GetPointsByUuid(workspaceId, ids)
	check(err)

	contexts := []model.ConsumableContext{}
	for _, point := range points {
		pointId := point.Id.GetUuid()
		payload := point.GetPayload()
		documentId := payload["documentId"].GetStringValue()
		text := payload["chunk"].GetStringValue()

		cc := model.ConsumableContext{ID: pointId, WorkspaceID: workspaceId, DocumentID: documentId, Text: text}
		contexts = append(contexts, cc)
	}

	if err == nil {
		res.JSON(http.StatusOK, contexts)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

// reCreate documentId
func (h Handler) ListChunks(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]
	documentId := req.Params["documentId"]
	pointId := req.String("pointId")
	adjacentRange := int64(2)

	// Get desired point
	point, err := h.QD.GetPoint(orgId, pointId)
	check(err)

	// extract document index from point
	payload := point.GetPayload()
	index := payload["index"].GetIntegerValue()

	// Get adjacent indeces and return all
	indeces := util.GetAdjacentIndeces(index, adjacentRange)

	points, err := h.QD.GetPointsByIndeces(orgId, documentId, indeces)

	chunks := []model.ConsumableContext{}
	for _, point := range points {
		pointId := point.Id.GetUuid()
		payload := point.GetPayload()
		documentId := payload["documentId"].GetStringValue()
		text := payload["chunk"].GetStringValue()

		cc := model.ConsumableContext{ID: pointId, WorkspaceID: workspaceId, DocumentID: documentId, Text: text}
		chunks = append(chunks, cc)
	}

	if err == nil {
		res.JSON(http.StatusOK, chunks)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetContext(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetContext(req.Params["contextId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteContext(res *goyave.Response, req *goyave.Request) {
	contextId := req.Params["contextId"]

	err := h.PG.DeleteContext(contextId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted context %s", contextId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Templates
//

func (h Handler) ListTemplates(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]

	results, err := h.PG.ListTemplates(orgId, workspaceId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateTemplate(res *goyave.Response, req *goyave.Request) {

	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject
	authorName, err := h.PG.GetUserName(userId)
	check(err)

	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]

	templateName := req.String("name")
	templateText := req.String("text")
	timestamp := time.Now().Format(time.RFC3339)

	result, err := h.PG.CreateTemplate(orgId, workspaceId, templateName, templateText, userId, authorName, timestamp)

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetTemplate(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetTemplate(req.Params["templateId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateTemplate(res *goyave.Response, req *goyave.Request) {
	templateId := req.Params["templateId"]
	templateName := req.String("name")
	templateText := req.String("text")

	result, err := h.PG.UpdateTemplate(templateId, templateName, templateText)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteTemplate(res *goyave.Response, req *goyave.Request) {
	templateId := req.Params["templateId"]

	err := h.PG.DeleteTemplate(templateId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted template %s", templateId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// Drive Docs (Google)
//

// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace/028c4bc1-9f6c-45b2-b975-84ffad13b852/drive
func (h Handler) ListDriveFolderDocs(res *goyave.Response, req *goyave.Request) {

	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject
	//fmt.Println("Listing", userId)

	query := req.Data["q"].(string)
	orderBy := req.Data["orderBy"].(string)
	limit := int64(req.Integer("limit"))
	provider := "oauth_google"

	token, status, err := h.checkToken(userId, provider)

	if err != nil {
		driveError := model.DriveError{Status: status, ErrorMessage: fmt.Sprintf("%s", err)}
		res.JSON(status, driveError)
		return
	}

	config, err := drive.GetDriveConfig()
	check(err)

	service, err := drive.GetDriveService(config, token)
	check(err)

	h.DRV = drive.Drv{Service: service}

	results, err := h.DRV.ListDocs(query, orderBy, limit)

	if err != nil { // happens if the user cancels their scopes after acquiring a valid oauth token
		gErr, ok := err.(*googleapi.Error)
		if ok {
			status := gErr.Code
			driveError := model.DriveError{Status: status, ErrorMessage: gErr.Message}
			res.JSON(status, driveError)
			return
		} else {
			status := http.StatusInternalServerError
			driveError := model.DriveError{Status: status, ErrorMessage: "unknown googleapi error"}
			res.JSON(status, driveError)
			return
		}
	}

	driveList := model.GoogleDriveFileList{ErrorStatus: false, Files: results}
	res.JSON(http.StatusOK, driveList)
}

func (h Handler) ListSharedDrives(res *goyave.Response, req *goyave.Request) {

	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject
	provider := "oauth_google"

	token, status, err := h.checkToken(userId, provider)

	if err != nil {
		driveError := model.DriveError{Status: status, ErrorMessage: fmt.Sprintf("%s", err)}
		res.JSON(status, driveError)
		return
	}

	config, err := drive.GetDriveConfig()
	fmt.Println("shared drives", config)
	check(err)

	// TODO: restore special check for drive services
	service, err := drive.GetDriveService(config, token)
	check(err) //driveCheck(err)

	//drv := Drv{Service: service} // add to Handler
	h.DRV = drive.Drv{Service: service}

	drives, err := h.DRV.ListSharedDrives()
	// checkDriveError(err)

	if err != nil { // happens if the user cancels their scopes after acquiring a valid oauth token
		gErr, ok := err.(*googleapi.Error)
		if ok {
			status := gErr.Code
			driveError := model.DriveError{Status: status, ErrorMessage: gErr.Message}
			res.JSON(status, driveError)
			return
		} else {
			status := http.StatusInternalServerError
			driveError := model.DriveError{Status: status, ErrorMessage: "unknown googleapi error"}
			res.JSON(status, driveError)
			return
		}
	}

	driveList := model.DriveList{ErrorStatus: false, Files: drives}
	res.JSON(http.StatusOK, driveList)
}

// generates report
// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace/2ce911d8-db4c-451c-8aa4-144794361dfb/drive/fcf2dd01-196b-4f3e-b409-867efe3a5a13
func (h Handler) GetSyncReport(res *goyave.Response, req *goyave.Request) {

	workspaceId := req.Params["workspaceId"]
	claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
	userId := claims.Subject
	provider := "oauth_google"

	token, status, err := h.checkToken(userId, provider)

	if err != nil {
		driveError := model.DriveError{Status: status, ErrorMessage: fmt.Sprintf("%s", err)}
		res.JSON(status, driveError)
		return
	}

	config, err := drive.GetDriveConfig()
	check(err)

	service, err := drive.GetDriveService(config, token)
	check(err)

	h.DRV = drive.Drv{Service: service}

	folderSyncs, err := h.PG.ListDriveFolderSync(workspaceId)
	check(err)

	documentSyncs, err := h.PG.ListDriveDocumentSync(workspaceId)
	check(err)

	driveSyncReport, err := h.createSyncReport(folderSyncs, documentSyncs)

	if err != nil {
		// can happen if:
		// user cancels scopes after acquiring valid oauth token
		// user does not have access to shared drive
		gErr, ok := err.(*googleapi.Error)
		if ok {
			status := gErr.Code
			driveError := model.DriveError{Status: status, ErrorMessage: gErr.Message}
			res.JSON(status, driveError)
			return
		} else {
			status := http.StatusInternalServerError
			driveError := model.DriveError{Status: status, ErrorMessage: "unknown sync report error"}
			res.JSON(status, driveError)
			return
		}
	} else {
		res.JSON(http.StatusOK, driveSyncReport)
	}
}

func (h Handler) GetSyncTree(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	folderSyncs, err := h.PG.ListDriveFolderSync(workspaceId)
	check(err)

	documentSyncs, err := h.PG.ListDriveDocumentSync(workspaceId)
	check(err)

	documents, err := h.PG.ListDocuments(workspaceId)
	check(err)

	parentageMap := util.CreateParentageMap(folderSyncs, documentSyncs)

	syncTree := model.SyncTree{
		FolderSyncs:   folderSyncs,
		DocumentSyncs: documentSyncs,
		Files:         documents,
		ParentageMap:  parentageMap,
	}

	if err == nil {
		res.JSON(http.StatusOK, syncTree)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteDriveFolderSync(res *goyave.Response, req *goyave.Request) {

	fmt.Println("Delete drive folder sync")

	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]
	syncId := req.Params["driveSyncId"]

	// Get folderSync entry
	folderSync, err := h.PG.GetDriveFolderSync(syncId)
	check(err)

	// Get all documentSyncs associated with folder
	documentSyncs, err := h.PG.ListDriveDocumentSyncByParentId(workspaceId, folderSync.DriveID)
	check(err)

	// Delete document vectors and documentSync entries
	for _, dSync := range documentSyncs {
		_, err = h.QD.DeleteVectorsByDocumentId(orgId, workspaceId, dSync.DocumentID)
		check(err)

		err = h.PG.DeleteDocument(dSync.DocumentID)
		check(err)

		err = h.PG.DeleteDriveDocumentSync(dSync.ID)
		check(err)
	}

	// Delete folderSync
	err = h.PG.DeleteDriveFolderSync(syncId)
	check(err)

	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
	if subscription.Active {
		record, err := h.createUsageEvent(orgId)
		fmt.Println(record)
		check(err)
	}

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted drive folder sync %s", syncId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// TAGS
//

func (h Handler) ListTags(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]

	results, err := h.PG.ListTags(orgId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) CreateTag(res *goyave.Response, req *goyave.Request) {
	orgId := req.Params["orgId"]
	tagName := req.String("name")
	tagColor := req.String("color")
	result, err := h.PG.CreateTag(orgId, tagName, tagColor)

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetTag(res *goyave.Response, req *goyave.Request) {
	result, err := h.PG.GetTemplate(req.Params["tagId"])

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UpdateTag(res *goyave.Response, req *goyave.Request) {
	tagId := req.Params["tagId"]
	tagName := req.String("name")
	tagColor := req.String("color")

	result, err := h.PG.UpdateTag(tagId, tagName, tagColor)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) DeleteTag(res *goyave.Response, req *goyave.Request) {
	tagId := req.Params["tagId"]
	err := h.PG.DeleteTag(tagId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted tag %s", tagId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListDocumentTagAssociations(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]

	results, err := h.PG.ListDocumentTagAssociations(workspaceId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) TagDocument(res *goyave.Response, req *goyave.Request) {
	workspaceId := req.Params["workspaceId"]
	documentId := req.Params["documentId"]
	tagId := req.Params["tagId"]

	result, err := h.PG.CreateDocumentTagAssociation(workspaceId, documentId, tagId)
	check(err)

	if err == nil {
		res.JSON(http.StatusCreated, result)
	} else {
		res.Status(http.StatusNotModified) // 304
		res.Error(err)
	}
}

func (h Handler) ListDocumentsByTagId(res *goyave.Response, req *goyave.Request) {
	tagId := req.Params["tagId"]

	results, err := h.PG.ListDocumentsByTagId(tagId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) ListTagsByDocumentId(res *goyave.Response, req *goyave.Request) {
	documentId := req.Params["documentId"]

	results, err := h.PG.ListTagsByDocumentId(documentId)

	if err == nil {
		res.JSON(http.StatusOK, results)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) GetDocumentTagAssociation(res *goyave.Response, req *goyave.Request) {
	documentId := req.Params["documentId"]
	tagId := req.Params["tagId"]

	result, err := h.PG.GetDocumentTagAssociation(documentId, tagId)

	if err == nil {
		res.JSON(http.StatusOK, result)
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

func (h Handler) UntagDocument(res *goyave.Response, req *goyave.Request) {
	documentId := req.Params["documentId"]
	tagId := req.Params["tagId"]

	err := h.PG.DeleteDocumentTagAssociation(documentId, tagId)

	if err == nil {
		res.JSON(http.StatusOK, model.HTTPResponse{Message: fmt.Sprintf("Deleted document-tag association %s %s", documentId, tagId)})
	} else {
		res.Status(http.StatusInternalServerError)
		res.Error(err)
	}
}

//
// ERROR HANDLING
//

func softCheck(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func check(err error) error {
	if err != nil {
		x := errors.New(err)
		fmt.Println(x.ErrorStack())
		return err
	}
	return nil
}
