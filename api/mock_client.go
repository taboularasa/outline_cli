package api

type MockClient struct {
	GetDocumentFunc    func(docID string, verbose bool) (*Document, error)
	UpdateDocumentFunc func(docID string, content string, verbose bool) error
	ListDocumentsFunc  func(verbose bool) ([]Document, error)
}

func (m *MockClient) GetDocument(docID string, verbose bool) (*Document, error) {
	return m.GetDocumentFunc(docID, verbose)
}

func (m *MockClient) UpdateDocument(docID string, content string, verbose bool) error {
	return m.UpdateDocumentFunc(docID, content, verbose)
}

func (m *MockClient) ListDocuments(verbose bool) ([]Document, error) {
	return m.ListDocumentsFunc(verbose)
}
