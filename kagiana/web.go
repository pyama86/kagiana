package kagiana

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/vault/sdk/helper/certutil"
	"github.com/sirupsen/logrus"
)

var header = `
<!DOCTYPE html>
<html lang="en" >
  <head>
    <meta charset="UTF-8">
    <title>Kagiana</title>
    <link rel='stylesheet' href='https://unpkg.com/bulma@0.9.0/css/bulma.min.css'>
    <script defer src="https://use.fontawesome.com/releases/v5.14.0/js/all.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/clipboard@2.0.6/dist/clipboard.min.js"></script>
    <style type="text/css">
.copy-value {
  cursor: pointer;
  position: relative;
}

.tooltip::after {
    content: 'Copied!';
    background: #555;
    display: inline-block;
    color: #fff;
    border-radius: .4rem;
    position: absolute;
    left: 50%;
    bottom: -.8rem;
    transform: translate(-50%, 0);
    font-size: .75rem;
    padding: 4px 10px 6px 10px;
    animation: fade-tooltip .5s 1s 1 forwards;
}

@keyframes fade-tooltip {
  to { opacity: 0; }
}
    </style>
  </head>
  <body>
    <section class="hero is-info">
      <div class="hero-body">
        <div class="columns">
          <div class="column is-12">
            <div class="container content">
              <h1 class="title">Kagiana</h1>
            </div>
          </div>
        </div>
      </div>
    </section>

`

var footer = `
    <footer class="footer">
      <hr>
      <div class="columns is-mobile is-centered">
        <div class="control">
          <div class="tags has-addons">
            <span class="tag is-dark">View on GitHub</span>
            <span class="tag has-addons is-warning"><a href="https://pyama86/kagiana"><i class="fab fa-lg fa-github"></i></a></span>
          </div>
        </div>
      </div>
      </div>
<script type="text/javascript">
const clipboard = new ClipboardJS(".copy-value");

const btns = document.querySelectorAll(".copy-value");

for(let i=0;i<btns.length;i++) {
    btns[i].addEventListener("mouseleave", clearTooltip);
}

function clearTooltip(e) {
    e.currentTarget.setAttribute("class","copy-value button is-info is-outlined mb-5");
}

function showTooltip(elem) {
    elem.setAttribute("class","copy-value tooltip button is-success is-outlined mb-5");
}

clipboard.on("success", function(e) {
    showTooltip(e.trigger);
});
</script>
    </footer>
  </body>
</html>
`
var errorTemplate = `
    <section class="section">
      <div class="container">
        <div class="columns">
          <div class="column">
            <div class="content is-medium">
              <h3 class="title is-3">Sorry...</h3>
              <div class="box">
                <h4 id="const" class="title is-3">
Status Code: {{ .StatusCode }}
                </h4>
                <article class="message">
                  <span class="icon has-text-danger">
                    <i class="fab"></i>
                  </span>
                  <div class="message-body">
{{ .Error }}
                  </div>
                </article>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
`
var successTemplate = `
    <section class="section">
      <div class="container">
        <div class="columns">
          <div class="column">
            <div class="content is-medium">
              <h3 class="title is-3">Vault Cert Get Successfly.</h3>
              <div class="box">
                <article class="message is-primary">
                  <span class="icon has-text-primary">
                    <i class="fab fa-info"></i>
                  </span>
                  <div class="message-body">
			Execute the following command.
                  </div>
                </article>
<button class="button is-primary is-outlined copy-value mb-5" data-clipboard-text="{{ .Command }}">
    Copy to clipboard
</button>
                <pre><code class="language-bash">
{{- range  $v := .MaskCommands }}
$ {{ $v -}}
{{- end -}}
                </code></pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
`

var echoContentPattern *regexp.Regexp = regexp.MustCompile(`".*"`)

func RenderSuccess(w http.ResponseWriter, cbs map[string]*certutil.CertBundle, token string) {
	commands := []string{
		`mkdir -p  ~/.kagiana`,
		fmt.Sprintf(`echo -e "%s" > ~/.kagiana/token`, token),
	}
	maskCommands := []string{
		`mkdir -p  ~/.kagiana`,
		`echo -e "*****" > ~/.kagiana/token`,
	}

	for name, cb := range cbs {
		for key, content := range map[string]string{
			"ca":   strings.Join(cb.CAChain, "\\n"),
			"cert": cb.Certificate,
			"key":  cb.PrivateKey,
		} {
			commands = append(commands, fmt.Sprintf(`echo -e "%s" > ~/.kagiana/%s.%s`, content, name, key))
			maskCommands = append(maskCommands, fmt.Sprintf(`echo -e "*****" > ~/.kagiana/%s.%s`, name, key))
		}

	}

	w.WriteHeader(http.StatusOK)
	tmpl, err := template.New("success").Parse(header + successTemplate + footer)
	if err != nil {
		logrus.Error(err)
	}
	err = tmpl.Execute(w, struct {
		MaskCommands []string
		Command      string
	}{
		MaskCommands: maskCommands,
		Command:      strings.Join(commands, ";\n"),
	})
	if err != nil {
		logrus.Error(err)
	}

	return
}

func RenderError(w http.ResponseWriter, statusCode int, displayError error) {
	w.WriteHeader(statusCode)
	tmpl, err := template.New("error").Parse(header + errorTemplate + footer)
	if err != nil {
		logrus.Error(err)
	}
	errString := "unknown error"
	if displayError != nil {
		logrus.Error(displayError)
		errString = displayError.Error()
	}
	err = tmpl.Execute(w, struct {
		StatusCode int
		Error      string
	}{
		StatusCode: statusCode,
		Error:      errString,
	})
	if err != nil {
		logrus.Error(err)
	}
}
