package htmlsimplifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimplifier_ProcessHTML_Footer(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
		{
			name: "footer with links and text",
			html: `
				<div class="col-lg-3 col-12 centered-lg">
					<p>
						<a href="https://www.nlm.nih.gov/web_policies.html" class="text-white">Web Policies</a><br>
						<a href="https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office" class="text-white">FOIA</a><br>
						<a href="https://www.hhs.gov/vulnerability-disclosure-policy/index.html" class="text-white" id="vdp">HHS Vulnerability Disclosure</a>
					</p>
				</div>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag:   "div",
				Attrs: "class=col-lg-3 col-12 centered-lg",
				Markdown: "[Web Policies](https://www.nlm.nih.gov/web_policies.html)\n" +
					"[FOIA](https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office)\n" +
					"[HHS Vulnerability Disclosure](https://www.hhs.gov/vulnerability-disclosure-policy/index.html)",
			},
		},
		{
			name: "footer with links and text - no markdown",
			html: `
				<div class="col-lg-3 col-12 centered-lg">
					<p>
						<a href="https://www.nlm.nih.gov/web_policies.html" class="text-white">Web Policies</a><br>
						<a href="https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office" class="text-white">FOIA</a><br>
						<a href="https://www.hhs.gov/vulnerability-disclosure-policy/index.html" class="text-white" id="vdp">HHS Vulnerability Disclosure</a>
					</p>
				</div>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     false,
			},
			expected: Document{
				Tag:   "div",
				Attrs: "class=col-lg-3 col-12 centered-lg",
				Children: []Document{
					{
						Tag:  "p",
						Text: "Web Policies\n\n\nFOIA\n\n\nHHS Vulnerability Disclosure",
					},
				},
			},
		},
		{
			name: "footer with mixed formatting",
			html: `
				<div class="col-lg-3 col-12 centered-lg">
					<p>
						<strong>Important:</strong> Please read our
						<a href="https://www.nlm.nih.gov/web_policies.html" class="text-white">Web Policies</a>
						and <em>privacy guidelines</em>.
					</p>
				</div>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag:   "div",
				Attrs: "class=col-lg-3 col-12 centered-lg",
				Markdown: "**Important:** Please read our " +
					"[Web Policies](https://www.nlm.nih.gov/web_policies.html) " +
					"and *privacy guidelines* .",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSimplifier(tt.opts)
			result, err := s.ProcessHTML(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimplifier_ProcessHTML_Lists(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
		// single li + a element should have markdown field
		{
			name: "single li + a element",
			html: `<li><a href="https://www.nlm.nih.gov/">NLM</a></li>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag:      "li",
				Markdown: "[NLM](https://www.nlm.nih.gov/)",
			},
		},
		{
			name: "navigation list with links",
			html: `
				<nav class="bottom-links">
					<ul class="mt-3">
						<li><a class="text-white" href="//www.nlm.nih.gov/">NLM</a></li>
						<li><a class="text-white" href="https://www.nih.gov/">NIH</a></li>
						<li><a class="text-white" href="https://www.hhs.gov/">HHS</a></li>
						<li><a class="text-white" href="https://www.usa.gov/">USA.gov</a></li>
					</ul>
				</nav>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag:   "nav",
				Attrs: "class=bottom-links",
				Children: []Document{
					{
						Tag:   "ul",
						Attrs: "class=mt-3",
						Children: []Document{
							{
								Tag:      "li",
								Markdown: "[NLM](//www.nlm.nih.gov/)",
							},
							{
								Tag:      "li",
								Markdown: "[NIH](https://www.nih.gov/)",
							},
							{
								Tag:      "li",
								Markdown: "[HHS](https://www.hhs.gov/)",
							},
							{
								Tag:      "li",
								Markdown: "[USA.gov](https://www.usa.gov/)",
							},
						},
					},
				},
			},
		},
		{
			name: "navigation list with max items",
			html: `
				<nav class="bottom-links">
					<ul class="mt-3">
						<li><a class="text-white" href="//www.nlm.nih.gov/">NLM</a></li>
						<li><a class="text-white" href="https://www.nih.gov/">NIH</a></li>
						<li><a class="text-white" href="https://www.hhs.gov/">HHS</a></li>
						<li><a class="text-white" href="https://www.usa.gov/">USA.gov</a></li>
					</ul>
				</nav>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
				MaxListItems: 2,
			},
			expected: Document{
				Tag:   "nav",
				Attrs: "class=bottom-links",
				Children: []Document{
					{
						Tag:   "ul",
						Attrs: "class=mt-3",
						Children: []Document{
							{
								Tag:      "li",
								Markdown: "[NLM](//www.nlm.nih.gov/)",
							},
							{
								Tag:      "li",
								Markdown: "[NIH](https://www.nih.gov/)",
							},
							{
								Text: "...",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSimplifier(tt.opts)
			result, err := s.ProcessHTML(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimplifier_ProcessHTML_SectionContainer(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
		{
			name: "section with container and multiple columns",
			html: `
				<section>
					<div class="container">
						<div class="row">
							<div class="col-lg-3 col-12 centered-lg">
								<p>
									<a href="https://www.nlm.nih.gov/web_policies.html" class="text-white">Web Policies</a><br>
									<a href="https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office" class="text-white">FOIA</a><br>
									<a href="https://www.hhs.gov/vulnerability-disclosure-policy/index.html" class="text-white" id="vdp">HHS Vulnerability Disclosure</a>
								</p>
							</div>
							<div class="col-lg-3 col-12 centered-lg">
								<p>
									<a class="supportLink text-white" href="https://support.nlm.nih.gov/">Help</a><br>
									<a href="https://www.nlm.nih.gov/accessibility.html" class="text-white">Accessibility</a><br>
									<a href="https://www.nlm.nih.gov/careers/careers.html" class="text-white">Careers</a>
								</p>
							</div>
						</div>
					</div>
				</section>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag: "section",
				Children: []Document{
					{
						Tag:   "div",
						Attrs: "class=container",
						Children: []Document{
							{
								Tag:   "div",
								Attrs: "class=row",
								Children: []Document{
									{
										Tag:   "div",
										Attrs: "class=col-lg-3 col-12 centered-lg",
										Markdown: "[Web Policies](https://www.nlm.nih.gov/web_policies.html)\n" +
											"[FOIA](https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office)\n" +
											"[HHS Vulnerability Disclosure](https://www.hhs.gov/vulnerability-disclosure-policy/index.html)",
									},
									{
										Tag:   "div",
										Attrs: "class=col-lg-3 col-12 centered-lg",
										Markdown: "[Help](https://support.nlm.nih.gov/)\n" +
											"[Accessibility](https://www.nlm.nih.gov/accessibility.html)\n" +
											"[Careers](https://www.nlm.nih.gov/careers/careers.html)",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSimplifier(tt.opts)
			result, err := s.ProcessHTML(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimplifier_ProcessHTML_CompleteDocument(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
		{
			name: "complete html document with doctype",
			html: `<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <div class="content">
        <p>Hello <a href="https://example.com">World</a></p>
    </div>
</body>
</html>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag: "body",
				Children: []Document{
					{
						Tag:      "div",
						Attrs:    "class=content",
						Markdown: "Hello [World](https://example.com)",
					},
				},
			},
		},
		{
			name: "complete html document with complex content",
			html: `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Complex Page</title>
    <meta charset="utf-8">
</head>
<body>
    <div class="container">
        <h1>Welcome</h1>
        <div class="row">
            <div class="col">
                <p>Visit our <a href="https://example.com/about">About</a> page</p>
            </div>
        </div>
    </div>
</body>
</html>`,
			opts: Options{
				SimplifyText: true,
				Markdown:     true,
			},
			expected: Document{
				Tag: "body",
				Children: []Document{
					{
						Tag:   "div",
						Attrs: "class=container",
						Children: []Document{
							{
								Tag:  "h1",
								Text: "Welcome",
							},
							{
								Tag:   "div",
								Attrs: "class=row",
								Children: []Document{
									{
										Tag:      "div",
										Attrs:    "class=col",
										Markdown: "Visit our [About](https://example.com/about) page",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSimplifier(tt.opts)
			result, err := s.ProcessHTML(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
