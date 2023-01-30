package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"webtool/pkg/logger"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/sync/errgroup"
)

func wafStart(ctx context.Context, s *Server, codes *ResponseCodes) error {
	var ls net.Listener
	// 建立socket监听
	ls, err := net.Listen("tcp", s.Addr)
	if err != nil {
		logger.Error().Msg(err.Error())
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		logger.Debug().
			Str("type", s.Type).
			Str("addr", s.Addr).
			Msg("strart")

		// 处理客户端连接
		for {
			conn, err := ls.Accept()
			if err != nil {
				logger.Error().Str("error", err.Error()).Msg("connect")
				return err
			}
			go tcpHandle(conn, codes)
		}
	})

	wg.Go(func() error {
		<-ctx.Done()
		if err := ls.Close(); err != nil {
			logger.Error().Msgf("Server:%s shutdown failed: %v", s.Addr, err)
			return fmt.Errorf("server:%s shutdown failed: %v", s.Addr, err)
		}
		logger.Error().Msg("Server Stopped")
		return nil
	})
	return wg.Wait()
}

func tcpHandle(conn net.Conn, codes *ResponseCodes) {
	defer conn.Close()
	addr := conn.RemoteAddr()
	logger.Error().Str("addr", addr.String()).Msgf("new connect -----------")

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for {
		isRequest, err := getPacket(reader, tp)

		// 说明tcp链接断了
		if err == io.EOF {
			logger.Error().Str("addr", addr.String()).Msgf("close connect -----------")
			return
		}
		logger.Debug().Msgf("handleConn isRequest:%v", isRequest)

		//响应体
		var respBody = "<h1>Hello World</h1>"
		length := len(respBody)

		//响应头
		// Status line
		var statusCode int
		if isRequest {
			statusCode = codes.RequestCode
		} else {
			statusCode = codes.ResponseCode
		}

		text := http.StatusText(statusCode)
		statusLine := fmt.Sprintf("HTTP/1.1 %03d %s\r\n", statusCode, text)

		var respHeader = statusLine +
			"Server: webtool\r\n" +
			"Content-Type: text/html;charset=ISO-8859-1\r\n" +
			"Content-Length: " + strconv.FormatInt(int64(length), 10)

		logger.Debug().Msgf("----> response statusLine:%s", statusLine)

		resp := respHeader + "\r\n\r\n" + respBody
		_, _ = conn.Write([]byte(resp))
	}
}

func getPacket(reader *bufio.Reader, tp *textproto.Reader) (isRequest bool, err error) {
	// 读第一首行
	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			return true, err
		}
		logger.Error().Str("err", err.Error()).Msg("read first line")
		return true, err
	}
	logger.Debug().Str("first line", line).Msg("getPacket")

	isRequest, req, resp, err := parsePacketLine(line)
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("parse first line")
		return true, err
	}

	// Subsequent lines: Key: value.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return isRequest, err
	}

	if isRequest {
		req.Header = http.Header(mimeHeader)
		err = readTransfer(req, reader)
	} else {
		resp.Header = http.Header(mimeHeader)
		err = readTransfer(resp, reader)
	}
	if err != nil {
		return isRequest, err
	}

	// 读body
	body := []byte{}
	if isRequest {
		logger.Debug().Int64("ContentLength", req.ContentLength).Msg("request")
		if req.ContentLength == 0 {
			return isRequest, nil
		}
		body = make([]byte, req.ContentLength)
		req.Body.Read(body)
	} else {
		logger.Debug().Int64("ContentLength", resp.ContentLength).Msg("response")
		if resp.ContentLength == 0 {
			return isRequest, nil
		}
		body = make([]byte, resp.ContentLength)
		resp.Body.Read(body)
	}
	logger.Debug().Str("Body", string(body)).Msg("read")
	return isRequest, nil
}

func parsePacketLine(line string) (isRequest bool, req *http.Request, resp *http.Response, err error) {
	req = new(http.Request)
	if "HTTP" == line[:4] {
		resp, err = parseResponseLine(line, req)
		if err != nil {
			return false, nil, nil, err
		}
		return false, req, resp, nil
	} else {
		// First line: GET /index.html HTTP/1.0
		var ok bool
		req.Method, req.RequestURI, req.Proto, ok = parseRequestLine(line)
		if !ok {
			return false, nil, nil, err
		}
		return true, req, nil, nil
	}
}

func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	method, rest, ok1 := strings.Cut(line, " ")
	requestURI, proto, ok2 := strings.Cut(rest, " ")
	if !ok1 || !ok2 {
		return "", "", "", false
	}
	return method, requestURI, proto, true
}

func parseResponseLine(line string, req *http.Request) (resp *http.Response, err error) {
	resp = &http.Response{
		Request: req,
	}
	proto, status, ok := strings.Cut(line, " ")
	if !ok {
		return nil, badStringError("malformed HTTP response", line)
	}
	resp.Proto = proto
	resp.Status = strings.TrimLeft(status, " ")

	statusCode, _, _ := strings.Cut(resp.Status, " ")
	if len(statusCode) != 3 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	if resp.ProtoMajor, resp.ProtoMinor, ok = http.ParseHTTPVersion(resp.Proto); !ok {
		return nil, badStringError("malformed HTTP version", resp.Proto)
	}

	return resp, nil
}

func badStringError(what, val string) error { return fmt.Errorf("%s %q", what, val) }

type transferReader struct {
	// Input
	Header        http.Header
	StatusCode    int
	RequestMethod string
	ProtoMajor    int
	ProtoMinor    int
	// Output
	Body          io.ReadCloser
	ContentLength int64
	Chunked       bool
	Close         bool
	Trailer       http.Header
}

// msg is *Request or *Response.
func readTransfer(msg any, r *bufio.Reader) (err error) {
	t := &transferReader{RequestMethod: "GET"}

	// Unify input
	isResponse := false
	switch rr := msg.(type) {
	case *http.Response:
		t.Header = rr.Header
		t.StatusCode = rr.StatusCode
		t.ProtoMajor = rr.ProtoMajor
		t.ProtoMinor = rr.ProtoMinor
		t.Close = shouldClose(t.ProtoMajor, t.ProtoMinor, t.Header, true)
		isResponse = true
		if rr.Request != nil {
			t.RequestMethod = rr.Request.Method
		}
	case *http.Request:
		t.Header = rr.Header
		t.RequestMethod = rr.Method
		t.ProtoMajor = rr.ProtoMajor
		t.ProtoMinor = rr.ProtoMinor
		// Transfer semantics for Requests are exactly like those for
		// Responses with status code 200, responding to a GET method
		t.StatusCode = 200
		t.Close = rr.Close
	default:
		panic("unexpected type")
	}

	// Default to HTTP/1.1
	if t.ProtoMajor == 0 && t.ProtoMinor == 0 {
		t.ProtoMajor, t.ProtoMinor = 1, 1
	}

	realLength, err := fixLength(isResponse, t.StatusCode, t.RequestMethod, t.Header, t.Chunked)
	if err != nil {
		return err
	}
	if isResponse && t.RequestMethod == "HEAD" {
		if n, err := parseContentLength(t.Header.Get("Content-Length")); err != nil {
			return err
		} else {
			t.ContentLength = n
		}
	} else {
		t.ContentLength = realLength
	}

	// Prepare body reader. ContentLength < 0 means chunked encoding
	// or close connection when finished, since multipart is not supported yet
	switch {
	case realLength == 0:
		t.Body = NoBody
	case realLength > 0:
		t.Body = &body{src: io.LimitReader(r, realLength), closing: t.Close}
	default:
		// realLength < 0, i.e. "Content-Length" not mentioned in header
		if t.Close {
			// Close semantics (i.e. HTTP/1.0)
			// t.Body = &body{src: r, closing: t.Close}
			t.Body = NoBody
		} else {
			// Persistent connection (i.e. HTTP/1.1)
			t.Body = NoBody
		}
	}

	// Unify output
	switch rr := msg.(type) {
	case *http.Request:
		rr.Body = t.Body
		rr.ContentLength = t.ContentLength
		if t.Chunked {
			rr.TransferEncoding = []string{"chunked"}
		}
		rr.Close = t.Close
		rr.Trailer = t.Trailer
	case *http.Response:
		rr.Body = t.Body
		rr.ContentLength = t.ContentLength
		if t.Chunked {
			rr.TransferEncoding = []string{"chunked"}
		}
		rr.Close = t.Close
		rr.Trailer = t.Trailer
	}

	return nil
}

func shouldClose(major, minor int, header http.Header, removeCloseHeader bool) bool {
	if major < 1 {
		return true
	}

	conv := header["Connection"]
	hasClose := httpguts.HeaderValuesContainsToken(conv, "close")
	if major == 1 && minor == 0 {
		return hasClose || !httpguts.HeaderValuesContainsToken(conv, "keep-alive")
	}

	if hasClose && removeCloseHeader {
		header.Del("Connection")
	}

	return hasClose
}

func parseContentLength(cl string) (int64, error) {
	cl = textproto.TrimString(cl)
	if cl == "" {
		return -1, nil
	}
	n, err := strconv.ParseUint(cl, 10, 63)
	if err != nil {
		return 0, badStringError("bad Content-Length", cl)
	}
	return int64(n), nil

}

func fixLength(isResponse bool, status int, requestMethod string, header http.Header, chunked bool) (int64, error) {
	isRequest := !isResponse
	contentLens := header["Content-Length"]

	// Hardening against HTTP request smuggling
	if len(contentLens) > 1 {
		// Per RFC 7230 Section 3.3.2, prevent multiple
		// Content-Length headers if they differ in value.
		// If there are dups of the value, remove the dups.
		// See Issue 16490.
		first := textproto.TrimString(contentLens[0])
		for _, ct := range contentLens[1:] {
			if first != textproto.TrimString(ct) {
				return 0, fmt.Errorf("http: message cannot contain multiple Content-Length headers; got %q", contentLens)
			}
		}

		// deduplicate Content-Length
		header.Del("Content-Length")
		header.Add("Content-Length", first)

		contentLens = header["Content-Length"]
	}

	// Logic based on response type or status
	// if noResponseBodyExpected(requestMethod) {
	// 	// For HTTP requests, as part of hardening against request
	// 	// smuggling (RFC 7230), don't allow a Content-Length header for
	// 	// methods which don't permit bodies. As an exception, allow
	// 	// exactly one Content-Length header if its value is "0".
	// 	if isRequest && len(contentLens) > 0 && !(len(contentLens) == 1 && contentLens[0] == "0") {
	// 		return 0, fmt.Errorf("http: method cannot contain a Content-Length; got %q", contentLens)
	// 	}
	// 	return 0, nil
	// }
	if status/100 == 1 {
		return 0, nil
	}
	switch status {
	case 204, 304:
		return 0, nil
	}

	// Logic based on Transfer-Encoding
	if chunked {
		return -1, nil
	}

	// Logic based on Content-Length
	var cl string
	if len(contentLens) == 1 {
		cl = textproto.TrimString(contentLens[0])
	}
	if cl != "" {
		n, err := parseContentLength(cl)
		if err != nil {
			return -1, err
		}
		return n, nil
	}
	header.Del("Content-Length")

	if isRequest {
		// RFC 7230 neither explicitly permits nor forbids an
		// entity-body on a GET request so we permit one if
		// declared, but we default to 0 here (not -1 below)
		// if there's no mention of a body.
		// Likewise, all other request methods are assumed to have
		// no body if neither Transfer-Encoding chunked nor a
		// Content-Length are set.
		return 0, nil
	}

	// Body-EOF logic based on other methods (like closing, or chunked coding)
	return -1, nil
}

var NoBody = noBody{}

type noBody struct{}

func (noBody) Read([]byte) (int, error)         { return 0, io.EOF }
func (noBody) Close() error                     { return nil }
func (noBody) WriteTo(io.Writer) (int64, error) { return 0, nil }

type body struct {
	src          io.Reader
	hdr          any           // non-nil (Response or Request) value means read trailer
	r            *bufio.Reader // underlying wire-format reader for the trailer
	closing      bool          // is the connection to be closed after reading body?
	doEarlyClose bool          // whether Close should stop early

	mu         sync.Mutex // guards following, and calls to Read and Close
	sawEOF     bool
	closed     bool
	earlyClose bool   // Close called and we didn't read to the end of src
	onHitEOF   func() // if non-nil, func to call when EOF is Read
}

// ErrBodyReadAfterClose is returned when reading a Request or Response
// Body after the body has been closed. This typically happens when the body is
// read after an HTTP Handler calls WriteHeader or Write on its
// ResponseWriter.
var ErrBodyReadAfterClose = errors.New("http: invalid Read on closed Body")

func (b *body) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, ErrBodyReadAfterClose
	}
	return b.readLocked(p)
}

// Must hold b.mu.
func (b *body) readLocked(p []byte) (n int, err error) {
	if b.sawEOF {
		return 0, io.EOF
	}
	n, err = b.src.Read(p)

	if err == io.EOF {
		b.sawEOF = true
		// Chunked case. Read the trailer.
		if b.hdr != nil {
			if e := b.readTrailer(); e != nil {
				err = e
				// Something went wrong in the trailer, we must not allow any
				// further reads of any kind to succeed from body, nor any
				// subsequent requests on the server connection. See
				// golang.org/issue/12027
				b.sawEOF = false
				b.closed = true
			}
			b.hdr = nil
		} else {
			// If the server declared the Content-Length, our body is a LimitedReader
			// and we need to check whether this EOF arrived early.
			if lr, ok := b.src.(*io.LimitedReader); ok && lr.N > 0 {
				err = io.ErrUnexpectedEOF
			}
		}
	}

	// If we can return an EOF here along with the read data, do
	// so. This is optional per the io.Reader contract, but doing
	// so helps the HTTP transport code recycle its connection
	// earlier (since it will see this EOF itself), even if the
	// client doesn't do future reads or Close.
	if err == nil && n > 0 {
		if lr, ok := b.src.(*io.LimitedReader); ok && lr.N == 0 {
			err = io.EOF
			b.sawEOF = true
		}
	}

	if b.sawEOF && b.onHitEOF != nil {
		b.onHitEOF()
	}

	return n, err
}

const maxPostHandlerReadBytes = 256 << 10

func (b *body) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	var err error
	switch {
	case b.sawEOF:
		// Already saw EOF, so no need going to look for it.
	case b.hdr == nil && b.closing:
		// no trailer and closing the connection next.
		// no point in reading to EOF.
	case b.doEarlyClose:
		// Read up to maxPostHandlerReadBytes bytes of the body, looking
		// for EOF (and trailers), so we can re-use this connection.
		if lr, ok := b.src.(*io.LimitedReader); ok && lr.N > maxPostHandlerReadBytes {
			// There was a declared Content-Length, and we have more bytes remaining
			// than our maxPostHandlerReadBytes tolerance. So, give up.
			b.earlyClose = true
		} else {
			var n int64
			// Consume the body, or, which will also lead to us reading
			// the trailer headers after the body, if present.
			n, err = io.CopyN(io.Discard, bodyLocked{b}, maxPostHandlerReadBytes)
			if err == io.EOF {
				err = nil
			}
			if n == maxPostHandlerReadBytes {
				b.earlyClose = true
			}
		}
	default:
		// Fully consume the body, which will also lead to us reading
		// the trailer headers after the body, if present.
		_, err = io.Copy(io.Discard, bodyLocked{b})
	}
	b.closed = true
	return err
}

type bodyLocked struct {
	b *body
}

func (bl bodyLocked) Read(p []byte) (n int, err error) {
	if bl.b.closed {
		return 0, ErrBodyReadAfterClose
	}
	return bl.b.readLocked(p)
}

var (
	singleCRLF = []byte("\r\n")
	doubleCRLF = []byte("\r\n\r\n")
)

var errTrailerEOF = errors.New("http: unexpected EOF reading trailer")

func seeUpcomingDoubleCRLF(r *bufio.Reader) bool {
	for peekSize := 4; ; peekSize++ {
		// This loop stops when Peek returns an error,
		// which it does when r's buffer has been filled.
		buf, err := r.Peek(peekSize)
		if bytes.HasSuffix(buf, doubleCRLF) {
			return true
		}
		if err != nil {
			break
		}
	}
	return false
}

func (b *body) readTrailer() error {
	// The common case, since nobody uses trailers.
	buf, err := b.r.Peek(2)
	if bytes.Equal(buf, singleCRLF) {
		b.r.Discard(2)
		return nil
	}
	if len(buf) < 2 {
		return errTrailerEOF
	}
	if err != nil {
		return err
	}

	// Make sure there's a header terminator coming up, to prevent
	// a DoS with an unbounded size Trailer. It's not easy to
	// slip in a LimitReader here, as textproto.NewReader requires
	// a concrete *bufio.Reader. Also, we can't get all the way
	// back up to our conn's LimitedReader that *might* be backing
	// this bufio.Reader. Instead, a hack: we iteratively Peek up
	// to the bufio.Reader's max size, looking for a double CRLF.
	// This limits the trailer to the underlying buffer size, typically 4kB.
	if !seeUpcomingDoubleCRLF(b.r) {
		return errors.New("http: suspiciously long trailer after chunked body")
	}

	hdr, err := textproto.NewReader(b.r).ReadMIMEHeader()
	if err != nil {
		if err == io.EOF {
			return errTrailerEOF
		}
		return err
	}
	switch rr := b.hdr.(type) {
	case *http.Request:
		mergeSetHeader(&rr.Trailer, http.Header(hdr))
	case *http.Response:
		mergeSetHeader(&rr.Trailer, http.Header(hdr))
	}
	return nil
}

func mergeSetHeader(dst *http.Header, src http.Header) {
	if *dst == nil {
		*dst = src
		return
	}
	for k, vv := range src {
		(*dst)[k] = vv
	}
}
