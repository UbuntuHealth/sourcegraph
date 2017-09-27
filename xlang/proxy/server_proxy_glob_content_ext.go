package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// handleTextDocumentContentExt handles textDocument/content requests
// adherent to the LSP files extension (see
// https://github.com/sourcegraph/language-server-protocol/pull/4).
func (c *serverProxyConn) handleTextDocumentContentExt(ctx context.Context, req *jsonrpc2.Request) (result interface{}, err error) {
	simulateFSLatency()

	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lspext.ContentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// Use package url not uri because this is a file:/// URI, not a
	// special Sourcegraph git://repo?rev#file URI.
	uri, err := url.Parse(string(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	if uri.Scheme != "file" {
		return nil, fmt.Errorf("textDocument/content only supports file: URIs (got %q)", uri)
	}

	contents, err := ctxvfs.ReadFile(ctx, c.rootFS, uri.Path)
	if err != nil {
		return nil, err
	}
	return &lsp.TextDocumentItem{Text: string(contents)}, nil
}

// handleWorkspaceFilesExt handles workspace/xfiles requests adherent to the
// LSP files extension (see
// https://github.com/sourcegraph/language-server-protocol/pull/4).
func (c *serverProxyConn) handleWorkspaceFilesExt(ctx context.Context, req *jsonrpc2.Request) (result interface{}, err error) {
	simulateFSLatency()

	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lspext.FilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// TODO(keegancsmith): Filter based on lspext.FilesParams.Base
	filenames, err := c.rootFS.ListAllFiles(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]lsp.TextDocumentIdentifier, 0, len(filenames))
	u := &url.URL{
		Scheme: "file",
	}
	for _, filename := range filenames {
		u.Path = "/" + filename
		res = append(res, lsp.TextDocumentIdentifier{URI: lsp.DocumentURI(u.String())})
	}

	return res, nil
}
