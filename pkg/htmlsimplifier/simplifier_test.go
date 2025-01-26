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
				Children: []Document{
					{
						Tag: "p",
						Markdown: "[Web Policies](https://www.nlm.nih.gov/web_policies.html)\n" +
							"[FOIA](https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office)\n" +
							"[HHS Vulnerability Disclosure](https://www.hhs.gov/vulnerability-disclosure-policy/index.html)",
					},
				},
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
						Tag: "p",
						Children: []Document{
							{
								Tag:   "a",
								Attrs: "href=https://www.nlm.nih.gov/web_policies.html class=text-white",
								Text:  "Web Policies",
							},
							{
								Tag: "br",
							},
							{
								Tag:   "a",
								Attrs: "href=https://www.nih.gov/institutes-nih/nih-office-director/office-communications-public-liaison/freedom-information-act-office class=text-white",
								Text:  "FOIA",
							},
							{
								Tag: "br",
							},
							{
								Tag:   "a",
								Attrs: "href=https://www.hhs.gov/vulnerability-disclosure-policy/index.html class=text-white id=vdp",
								Text:  "HHS Vulnerability Disclosure",
							},
						},
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
				Children: []Document{
					{
						Tag: "p",
						Markdown: "**Important:** Please read our " +
							"[Web Policies](https://www.nlm.nih.gov/web_policies.html) " +
							"and *privacy guidelines*.",
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

func TestSimplifier_ProcessHTML_Lists(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
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
