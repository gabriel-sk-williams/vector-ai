package route

import (
	bg "context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"vector-ai/drive"
	"vector-ai/model"
	"vector-ai/parse"
	"vector-ai/util"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscriptionitem"
	"github.com/stripe/stripe-go/v76/usagerecord"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/textsplitter"
)

type Options struct {
	VectorSize   int
	ChunkSize    int
	ChunkOverlap int
}

func (h Handler) parseLocalUpload(evs model.EventStream, nlp model.NewLocalProfile) (model.EventStream, string) {

	header := nlp.CoreDocumentProps
	workspaceId := nlp.WorkspaceID
	documentId := nlp.DocumentID

	var event model.UploadEvent
	event = h.broadcast("Parsing", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	parsedDoc, err := parse.ParseLocal(header, nlp.Data)

	event = h.broadcast("Parsing", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	return evs, parsedDoc
}

func (h Handler) splitEmbedUpload(evs model.EventStream, vsp model.VectorStorageProfile, parsedDoc string, embedder *embeddings.EmbedderImpl, opt Options) (model.EventStream, int64) {

	orgId := vsp.OrgID
	workspaceId := vsp.WorkspaceID
	documentId := vsp.DocumentID

	options := []textsplitter.Option{
		textsplitter.WithSeparators([]string{"\n\n", "\n", " ", ""}),
		textsplitter.WithChunkOverlap(opt.ChunkOverlap),
		textsplitter.WithChunkSize(opt.ChunkSize),
		// textsplitter.WithLenFunc(func(s string) int { return len(tokenEncoder.Encode(s, nil, nil)) }),
	}

	rc := textsplitter.NewRecursiveCharacter(options...)

	var event model.UploadEvent
	event = h.broadcast("Splitting", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	// fmt.Println("parsed", parsedDoc)
	chunks, err := rc.SplitText(parsedDoc)

	event = h.broadcast("Splitting", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)
	event = h.broadcast("Embedding", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	ctx := bg.Background()
	floats, err := embedder.EmbedDocuments(ctx, chunks)
	ctx.Done()

	event = h.broadcast("Embedding", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)
	event = h.broadcast("Uploading", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	result, err := h.QD.Upload(orgId, workspaceId, documentId, floats, chunks)
	fmt.Println(result)

	event = h.broadcast("Uploading", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	return evs, int64(len(chunks))
}

func (h Handler) saveLocalDocument(evs model.EventStream, nlp model.NewLocalProfile, chunks int64, opt Options) model.EventStream {

	uuid := nlp.ID
	fileName := nlp.Name
	fileSize := nlp.Size
	mimeType := nlp.MimeType
	workspaceId := nlp.WorkspaceID
	documentId := nlp.DocumentID

	var event model.UploadEvent
	event = h.broadcast("Updating", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	timestamp := time.Now().Format(time.RFC3339)
	doc, err := h.PG.CreateDocument(uuid, workspaceId, fileName, mimeType, fileSize, chunks, int64(opt.ChunkSize), timestamp)
	fmt.Println(doc)

	event = h.broadcast("Updating", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	h.broadcast("Operation", "Completed", workspaceId, documentId, err)

	return evs
}

func (h Handler) downloadDriveFile(evs model.EventStream, dp model.DownloadProfile) (model.EventStream, io.ReadCloser, string) {

	driveId := dp.DriveID
	mimeType := dp.MimeType
	workspaceId := dp.WorkspaceID
	documentId := dp.DocumentID

	var event model.UploadEvent
	var res *http.Response
	var err error

	exportType := parse.GoogleDriveExportType(mimeType)
	if mimeType == "application/vnd.google-apps.document" {

		event = h.broadcast("Exporting", "Started", workspaceId, documentId, nil)
		evs.Events = append(evs.Events, event)

		// Message: "Export only supports Docs Editors files."
		res, err = h.DRV.ExportDriveFile(driveId, exportType)
		check(err)

		event = h.broadcast("Exporting", "Completed", workspaceId, documentId, err)
		evs.Events = append(evs.Events, event)

	} else {

		event = h.broadcast("Downloading", "Started", workspaceId, documentId, nil)
		evs.Events = append(evs.Events, event)

		res, err = h.DRV.DownloadDriveFile(driveId)
		check(err)

		event = h.broadcast("Downloading", "Completed", workspaceId, documentId, err)
		evs.Events = append(evs.Events, event)
	}

	return evs, res.Body, exportType
}

func (h Handler) parseBody(evs model.EventStream, md model.ManifestData, body io.ReadCloser, exportType string) (model.EventStream, string) {
	workspaceId := md.WorkspaceID
	documentId := md.DocumentID
	fileSize := md.Size

	var event model.UploadEvent
	event = h.broadcast("Parsing", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	parsedDoc, err := parse.BodyText(body, fileSize, exportType)

	event = h.broadcast("Parsing", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	return evs, parsedDoc
}

func (h Handler) syncNew(evs model.EventStream, ndp model.NewDriveProfile, chunks int64, opt Options) model.EventStream {

	uuid := ndp.ID
	documentId := uuid.String()
	workspaceId := ndp.WorkspaceID
	driveId := ndp.DriveID
	parentId := ndp.DriveParentID
	fileName := ndp.Name
	mimeType := ndp.MimeType
	fileSize := ndp.Size
	lastModified := ndp.LastModified.Format(time.RFC3339)

	var event model.UploadEvent
	event = h.broadcast("Updating", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	timestamp := time.Now().Format(time.RFC3339)
	_, err := h.PG.CreateDocument(uuid, workspaceId, fileName, mimeType, fileSize, chunks, int64(opt.ChunkSize), timestamp)

	event = h.broadcast("Updating", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)
	event = h.broadcast("Synchronizing", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	_, err = h.PG.CreateDriveDocumentSync(workspaceId, documentId, driveId, parentId, drive.ServiceType, lastModified)

	event = h.broadcast("Synchronizing", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	h.broadcast("Operation", "Completed", workspaceId, documentId, err)

	return evs
}

func (h Handler) syncUpdated(evs model.EventStream, udp model.UpdatedDriveProfile, chunks int64, opt Options) model.EventStream {

	syncId := udp.SyncID
	documentId := udp.DocumentID
	workspaceId := udp.WorkspaceID
	fileName := udp.Name
	fileSize := udp.Size
	lastModified := udp.LastModified.Format(time.RFC3339)

	var event model.UploadEvent
	event = h.broadcast("Updating", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	timestamp := time.Now().Format(time.RFC3339)
	_, err := h.PG.UpdateDocument(documentId, fileName, int64(fileSize), chunks, int64(opt.ChunkSize), timestamp)

	event = h.broadcast("Updating", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)
	event = h.broadcast("Synchronizing", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	// syncId comes from ID of drive_document_sync
	_, err = h.PG.UpdateDriveDocumentSyncLastModified(syncId, lastModified)

	h.broadcast("Synchronizing", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	h.broadcast("Operation", "Completed", workspaceId, documentId, err)

	return evs
}

func (h Handler) syncMissing(evs model.EventStream, mdp model.MissingDriveProfile) model.EventStream {

	syncId := mdp.SyncID
	documentId := mdp.DocumentID
	workspaceId := mdp.WorkspaceID
	fileName := mdp.Name
	fileSize := mdp.Size
	fmt.Printf("Deleting missing file: %+v\n", fileName)
	fmt.Printf("File Size: %+v\n", fileSize)

	var event model.UploadEvent
	event = h.broadcast("Updating", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	err := h.PG.DeleteDocument(documentId)

	event = h.broadcast("Updating", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)
	event = h.broadcast("Synchronizing", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	err = h.PG.DeleteDriveDocumentSync(syncId)

	event = h.broadcast("Synchronizing", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	h.broadcast("Operation", "Completed", workspaceId, documentId, err)

	return evs
}

func (h Handler) DeleteVectors(evs model.EventStream, vsp model.VectorStorageProfile) model.EventStream {
	orgId := vsp.OrgID
	workspaceId := vsp.WorkspaceID
	documentId := vsp.DocumentID

	var event model.UploadEvent
	event = h.broadcast("Deleting", "Started", workspaceId, documentId, nil)
	evs.Events = append(evs.Events, event)

	pointsDeleted, err := h.QD.DeleteVectorsByDocumentId(orgId, workspaceId, documentId)
	detail := fmt.Sprintf("Deleted %d points", pointsDeleted)
	fmt.Println(detail)

	event = h.broadcast("Deleting", "Completed", workspaceId, documentId, err)
	evs.Events = append(evs.Events, event)

	return evs
}

func (h Handler) broadcast(op string, action string, workspaceId string, documentId string, err error) model.UploadEvent {
	var event model.UploadEvent
	if err == nil {
		progress := progress(op)
		event = model.UploadEvent{Operation: op, Action: action}
		h.TR.Broadcast(model.UploadStatus(event, workspaceId, documentId, progress))
	} else {
		event = model.UploadEvent{Operation: op, Action: "Failed", Detail: err.Error()}
		h.TR.Broadcast(model.ErrorStatus(event, workspaceId, documentId, 100))
	}
	return event
}

func progress(status string) int {

	// Manual:
	// Opening - Parsing
	// Splitting - Embedding - Uploading
	// Updating

	// Drive Sync:
	// Downloading/Exporting - Parsing
	// Splitting - Embedding - Uploading
	// Updating - Synchronizing

	if status == ("Opening") {
		return 10
	} else if status == ("Downloading") || status == ("Exporting") {
		return 10
	} else if status == ("Parsing") {
		return 15
	} else if status == ("Splitting") {
		return 30
	} else if status == ("Embedding") {
		return 45
	} else if status == ("Uploading") {
		return 60
	} else if status == ("Updating") {
		return 80
	} else if status == ("Synchronizing") {
		return 90
	} else if status == ("Operation") {
		return 100
	} else {
		return 100
	}
}

func (h Handler) createUsageEvent(orgId string) (*stripe.UsageRecord, error) {
	// usage events
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	totalFileSizeAmount, err := h.PG.GetTotalFileSizeAmount(orgId)
	check(err)
	_, err = h.PG.CreateUsageEvent(orgId, totalFileSizeAmount, time.Now().Format(time.RFC3339))
	check(err) // broadcast

	fmt.Println("tfsa", totalFileSizeAmount)

	subscription, err := h.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
	fmt.Println(subscription)
	check(err)
	stripeId := subscription.StripeSubscriptionID

	mb := util.ConvertBytesToMB(totalFileSizeAmount)
	fmt.Println("mbs", mb)

	params := &stripe.SubscriptionItemListParams{
		Subscription: stripe.String(stripeId),
	}
	params.Limit = stripe.Int64(1)
	result := subscriptionitem.List(params)

	var si *stripe.SubscriptionItem
	for result.Next() {
		si = result.SubscriptionItem()
		break
	}

	urParams := &stripe.UsageRecordParams{
		Quantity:         stripe.Int64(mb),
		SubscriptionItem: stripe.String(si.ID),
	}

	return usagerecord.New(urParams)
}
