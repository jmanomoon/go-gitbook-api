package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/GitbookIO/go-gitbook-api/client"
	"github.com/GitbookIO/go-gitbook-api/models"
	"github.com/GitbookIO/go-gitbook-api/streams"

	"mime/multipart"
)

type Book struct {
	Client *client.Client
}

type postStream func(bookId, version, branch string, r io.Reader) error

// Get returns a books details for a given "bookId"
// (for example "gitbookio/javascript")
func (b *Book) Get(bookId string) (models.Book, error) {
	book := models.Book{}

	_, err := b.Client.Get(
		fmt.Sprintf("/book/%s", bookId),
		nil,
		&book,
	)

	return book, err
}

// Build should only be used by internal clients, Publish by others
// Build starts a build and will not update the backing git repository
func (b *Book) Build(bookId, version, branch, bookpath string) error {
	return b.doStreamPublish(bookId, version, branch, bookpath, streams.PickStream, b.PublishBuildStream)
}

// PublishGit packages a git repo as tar.gz and uploads it to gitbook.io
func (b *Book) BuildGit(bookId, version, branch, bookpath, ref string) error {
	return b.doStreamPublish(bookId, version, branch, bookpath, streams.GitRef(ref), b.PublishBuildStream)
}

// PublishFolder packages a folder as tar.gz and uploads it to gitbook.io
func (b *Book) BuildFolder(bookId, version, branch, bookpath string) error {
	return b.doStreamPublish(bookId, version, branch, bookpath, streams.Folder, b.PublishBuildStream)
}

// PublishTarGz publishes a book based on a tar.gz file
func (b *Book) BuildTarGz(bookId, version, branch, bookpath string) error {
	return b.doStreamPublish(bookId, version, branch, bookpath, streams.File, b.PublishBuildStream)
}

func (b *Book) doStreamPublish(bookId, version, branch, bookpath string, streamfn streams.StreamFunc, postfn postStream) error {
	stream, err := streamfn(bookpath)
	if err != nil {
		return err
	}
	defer stream.Close()

	return postfn(bookId, version, branch, stream)
}

func (b *Book) PublishBuildStream(bookId, version, branch string, r io.Reader) error {
	return b.PublishStream(
		fmt.Sprintf("/book/%s/build/%s", bookId, version),
		version,
		branch,
		r,
	)
}

// PublishStream
func (b *Book) PublishStream(_url, version, branch string, r io.Reader) error {
	// Build request
	req, err := newfileUploadRequest(
		b.Client.Url(_url),
		// No params
		nil,
		"book",
		r,
	)
	if err != nil {
		return err
	}

	uinfo := b.Client.Userinfo

	// Auth
	pwd, _ := uinfo.Password()
	req.SetBasicAuth(uinfo.Username(), pwd)

	// Set version
	values := url.Values{}
	values.Set("version", version)
	values.Set("branch", branch)
	req.URL.RawQuery = values.Encode()

	// Execute request
	response, err := b.Client.Client.Do(req)
	if err != nil {
		return err
	}
	// Close body immediately to avoid leaks
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf(string(data[:]))
	}

	// Some error to code
	if response.StatusCode >= 400 {
		errMsg, err := client.DecodeError(response.Body)
		if err != nil {
			return err
		}
		return errMsg
	}

	return nil
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName string, reader io.Reader) (*http.Request, error) {
	// Buffer for body
	body := &bytes.Buffer{}
	// Multipart data
	writer := multipart.NewWriter(body)

	// File part
	part, err := writer.CreateFormFile(paramName, "book.tar.gz")
	if err != nil {
		return nil, err
	}

	// Copy over data for file
	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, err
	}

	// Write extra fields
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", uri, body)
	if err != nil {
		return nil, err
	}

	// Set header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
