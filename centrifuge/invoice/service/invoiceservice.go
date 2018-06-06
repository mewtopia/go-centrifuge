package invoiceservice

import (
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentService struct{}

// HandleAnchorInvoiceDocument anchors the given invoice document and returns the anchor details
func (s *InvoiceDocumentService) HandleAnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *invoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	err := invoicerepository.GetInvoiceRepository().Store(anchorInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	inv := invoice.NewInvoice(anchorInvoiceEnvelope.Document)
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()
	// Signing of document missing so far


	err = coreDoc.Anchor()
	if err != nil {
		return nil, err
	}

	return anchorInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) HandleSendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	err := invoicerepository.GetInvoiceRepository().Store(sendInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	inv := invoice.NewInvoice(sendInvoiceEnvelope.Document)
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()
	// Sign document
	// Uncomment once fixed
	//coreDoc.Sign()
	//coreDoc.Anchor()

	errs := []error{}
	for _, element := range sendInvoiceEnvelope.Recipients {
		err1 := coreDoc.Send(ctx, string(element[:]))
		if err1 != nil {
			errs = append(errs, err1)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, fmt.Errorf("%v", errs)
	}
	return sendInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) HandleGetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := invoicerepository.GetInvoiceRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		doc1, err1 := coredocumentrepository.GetCoreDocumentRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
		if err1 == nil {
			doc = invoice.NewInvoiceFromCoreDocument(&coredocument.CoreDocument{doc1}).Document
			err = err1
		}
		log.Errorf("%v", err)
	}
	return doc, err
}

func (s *InvoiceDocumentService) HandleGetReceivedInvoiceDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return nil, nil
}