package api

type MockClient struct {
	GetDocumentFunc    func(docID string) (*Document, error)
	UpdateDocumentFunc func(docID string, content string) error
}

func (m *MockClient) GetDocument(docID string) (*Document, error) {
	return m.GetDocumentFunc(docID)
}

func (m *MockClient) UpdateDocument(docID string, content string) error {
	return m.UpdateDocumentFunc(docID, content)
}
