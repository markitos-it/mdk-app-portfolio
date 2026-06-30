package main

import (
	"fmt"
	"io/fs"
	"log"
	portfolio "markitos-it-app-personal-portfolio"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	clients "github.com/markitos-it/mdk-event-relay-client/client"
)

var pageRoutes = map[string]string{
	"/":         "index.html",
	"/about":    "about.html",
	"/projects": "projects.html",
	"/resume":   "resume.html",
	"/works":    "works.html",
	"/blog":     "blog.html",
	"/contact":  "contact.html",
	"/register": "register.html",
}

type app struct{}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.bytesWritten += len(data)
	return rw.ResponseWriter.Write(data)
}

var eventClientCalledMaximum = 10
var eventClientCalledCurrent = 0
var eventRelay = clients.NewEventRelayClient(os.Getenv("EVENT_RELAY_DATABASE_PATH"))

func main() {
	app := app{}

	validateEnvVars()

	assetsFS, err := fs.Sub(portfolio.FS, "assets")
	if err != nil {
		log.Fatalf("Error creating sub-filesystem for assets: %v", err)
	}

	http.Handle("/assets/", app.loggingMiddleware(http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS)))))
	http.HandleFunc("/contact/submit", app.loggingMiddleware(http.HandlerFunc(app.handleContactSubmit)).ServeHTTP)
	http.HandleFunc("/register/submit", app.loggingMiddleware(http.HandlerFunc(app.handleRegisterSubmit)).ServeHTTP)
	http.HandleFunc("/", app.loggingMiddleware(http.HandlerFunc(app.handleRequest)).ServeHTTP)
	http.HandleFunc("/event-client/say-hello", app.loggingMiddleware(http.HandlerFunc(app.handleEventClientSayHelloRequest)).ServeHTTP)

	serverAddress := getServerAddress()
	log.Printf("Server started at http://%s\n", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func validateEnvVars() {
	if os.Getenv("EVENT_RELAY_DATABASE_PATH") == "" {
		log.Fatal("EVENT_RELAY_DATABASE_PATH environment variable is not set")
	}
	if os.Getenv("SERVER_ADDRESS") == "" {
		log.Fatal("SERVER_ADDRESS environment variable is not set")
	}
}

func (a app) loggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		handler.ServeHTTP(rw, r)

		duration := time.Since(start)
		a.logAccess(r, rw.statusCode, rw.bytesWritten, duration)
	})
}

func (a app) logAccess(r *http.Request, statusCode, bytesWritten int, duration time.Duration) {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "-"
	}
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "-"
	}

	logEntry := fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" (%.3fms)",
		clientIP,
		timestamp,
		r.Method,
		r.RequestURI,
		r.Proto,
		statusCode,
		bytesWritten,
		referer,
		userAgent,
		duration.Seconds()*1000,
	)

	log.Println(logEntry)
}

func (a app) handleEventClientSayHelloRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	eventRelay.Publish("say-hello-have-been-requested", `{
		"event": "hay-hello", 
		"name": "markitos"
	}`)

	eventClientCalledCurrent = eventClientCalledCurrent + 1
	if eventClientCalledCurrent > eventClientCalledMaximum {
		http.Error(w, "Too Many Say Hello =:()", http.StatusTooManyRequests)
		return
	}

	http.Redirect(w, r, "/?say-hello=enga", http.StatusSeeOther)
}

func (a app) handleRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	repassword := r.Form.Get("repassword")

	if repassword != password {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/register"
	}
	u, err := url.Parse(referer)
	if err != nil {
		log.Printf("Error parsing referer URL: %v", err)
		http.Redirect(w, r, referer, http.StatusSeeOther)
	}

	payload := fmt.Sprintf(`{"event": "register", "name": "%s", "email": "%s"}`, name, email)
	status := "success"
	eventRelay.Publish("register-have-been-created", payload)

	q := u.Query()
	q.Set("status", status)
	u.RawQuery = q.Encode()
	referer = u.String()

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func (a app) handleContactSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	email := r.Form.Get("email")
	message := r.Form.Get("message")
	name = template.HTMLEscapeString(name)
	email = template.HTMLEscapeString(email)
	message = template.HTMLEscapeString(message)

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/contact"
	}
	u, err := url.Parse(referer)
	if err != nil {
		log.Printf("Error parsing referer URL: %v", err)
		http.Redirect(w, r, referer, http.StatusSeeOther)
	}

	payload := fmt.Sprintf(`{"event": "contact", "name": "%s", "email": "%s", "message": "%s"}`,
		name, email, strings.ReplaceAll(strings.ReplaceAll(message, "\n", "\\n"), "\r", "\\r"))
	eventRelay.Publish("contact-have-been-sended", payload)

	status := "success"
	q := u.Query()
	q.Set("status", status)
	u.RawQuery = q.Encode()
	referer = u.String()

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func (a app) handleRequest(w http.ResponseWriter, r *http.Request) {
	lang, reqPath := detectLanguageAndPath(r.URL.Path)

	if reqPath != "/" && strings.HasSuffix(r.URL.Path, "/") {
		redirectPath := reqPath
		if lang == "es" {
			redirectPath = "/es" + reqPath
		}
		http.Redirect(w, r, redirectPath, http.StatusMovedPermanently)
		return
	}

	filename, ok := pageRoutes[reqPath]
	if !ok {
		http.NotFound(w, r)
		return
	}

	a.renderPage(w, r, filename, reqPath, lang)
}

func detectLanguageAndPath(rawPath string) (lang, reqPath string) {
	lang = "en"
	reqPath = strings.TrimSuffix(rawPath, "/")

	if strings.HasPrefix(reqPath, "/es") {
		lang = "es"
		reqPath = strings.TrimPrefix(reqPath, "/es")
	}

	if reqPath == "" {
		reqPath = "/"
	}

	return lang, reqPath
}

func (a app) renderPage(w http.ResponseWriter, r *http.Request, filename, reqPath, lang string) {
	tmpl := template.New("")

	langTemplatesDir, langPagesDir := a.languageDirs(lang)
	if err := loadTemplates(tmpl, langTemplatesDir, lang); err != nil {
		log.Printf("Error loading templates: %v", err)
		http.Error(w, "Error loading templates", http.StatusInternalServerError)
		return
	}

	if err := parsePage(tmpl, path.Join(langPagesDir, filename), filename); err != nil {
		log.Printf("Error loading page: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Path":              reqPath,
		"Lang":              lang,
		"ShowSayHelloToast": r.URL.Query().Get("say-hello") != "",
	}
	if err := tmpl.ExecuteTemplate(w, filename, data); err != nil {
		log.Printf("Error rendering page: %v", err)
	}
}

func (a app) languageDirs(lang string) (templatesDir, pagesDir string) {
	templatesDir = "templates"
	pagesDir = "pages"

	if lang == "es" {
		templatesDir = path.Join(templatesDir, "es")
		pagesDir = path.Join(pagesDir, "es")
	}
	if lang == "en" {
		templatesDir = path.Join(templatesDir, "en")
		pagesDir = path.Join(pagesDir, "en")
	}

	return templatesDir, pagesDir
}

func loadTemplates(tmpl *template.Template, templatesDir, lang string) error {
	entries, err := portfolio.FS.ReadDir(templatesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		tfile := path.Join(templatesDir, entry.Name())
		content, err := portfolio.FS.ReadFile(tfile)
		if err != nil {
			return err
		}

		templateName := templateNameFor(lang, tfile)
		if _, err := tmpl.New(templateName).Parse(string(content)); err != nil {
			return err
		}
	}

	return nil
}

func parsePage(tmpl *template.Template, pagePath, filename string) error {
	pageContent, err := portfolio.FS.ReadFile(pagePath)
	if err != nil {
		return err
	}

	_, err = tmpl.New(filename).Parse(string(pageContent))
	return err
}

func templateNameFor(lang, tfile string) string {
	prefix := "templates"
	if lang == "es" {
		prefix = "templates/es"
	}
	if lang == "en" {
		prefix = "templates/en"
	}

	return path.Join(prefix, path.Base(tfile))
}

func getServerAddress() string {
	result := os.Getenv("SERVER_ADDRESS")
	if result == "" {
		result = ":3080"
	}
	return result
}
