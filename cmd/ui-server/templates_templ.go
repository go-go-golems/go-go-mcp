// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package main

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

func base(title string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 1)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 15, Col: 17}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 2)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 3)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func indexTemplate(pages map[string]UIDefinition) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var4 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 4)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for name := range pages {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 5)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 templ.SafeURL = templ.SafeURL("/" + strings.TrimSuffix(name, ".yaml"))
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var5)))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 6)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(name)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 103, Col: 13}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 7)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 8)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = base("UI Server - Index").Render(templ.WithChildren(ctx, templ_7745c5c3_Var4), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func pageTemplate(name string, def UIDefinition) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var7 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var7 == nil {
			templ_7745c5c3_Var7 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var8 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 9)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, component := range def.Components {
				for typ, props := range component {
					templ_7745c5c3_Err = renderComponent(typ, props.(map[string]interface{})).Render(ctx, templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 10)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var9 string
			templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(yamlString(def))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 135, Col: 56}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 11)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = base("UI Server - "+name).Render(templ.WithChildren(ctx, templ_7745c5c3_Var8), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func renderComponent(typ string, props map[string]interface{}) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var10 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var10 == nil {
			templ_7745c5c3_Var10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		switch typ {
		case "button":
			if id, ok := props["id"].(string); ok {
				var templ_7745c5c3_Var11 = []any{
					"btn",
					templ.KV("btn-primary", props["type"] == "primary"),
					templ.KV("btn-secondary", props["type"] == "secondary"),
					templ.KV("btn-danger", props["type"] == "danger"),
					templ.KV("btn-success", props["type"] == "success"),
				}
				templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var11...)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 12)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var12 string
				templ_7745c5c3_Var12, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 148, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var12))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 13)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var13 string
				templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("logToConsole('%s clicked')", id))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 149, Col: 68}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 14)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if disabled, ok := props["disabled"].(bool); ok && disabled {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 15)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 16)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var14 string
				templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var11).String())
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 1, Col: 0}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 17)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if text, ok := props["text"].(string); ok {
					var templ_7745c5c3_Var15 string
					templ_7745c5c3_Var15, templ_7745c5c3_Err = templ.JoinStringErrs(text)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 162, Col: 11}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var15))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 18)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		case "title":
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 19)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 20)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var16 string
				templ_7745c5c3_Var16, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 169, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var16))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 21)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 22)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if content, ok := props["content"].(string); ok {
				var templ_7745c5c3_Var17 string
				templ_7745c5c3_Var17, templ_7745c5c3_Err = templ.JoinStringErrs(content)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 173, Col: 13}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var17))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 23)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		case "text":
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 24)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 25)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var18 string
				templ_7745c5c3_Var18, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 179, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var18))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 26)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 27)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if content, ok := props["content"].(string); ok {
				var templ_7745c5c3_Var19 string
				templ_7745c5c3_Var19, templ_7745c5c3_Err = templ.JoinStringErrs(content)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 183, Col: 13}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var19))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 28)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		case "input":
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 29)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 30)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var20 string
				templ_7745c5c3_Var20, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 189, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var20))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 31)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if typ, ok := props["type"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 32)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var21 string
				templ_7745c5c3_Var21, templ_7745c5c3_Err = templ.JoinStringErrs(typ)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 192, Col: 14}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var21))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 33)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if placeholder, ok := props["placeholder"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 34)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var22 string
				templ_7745c5c3_Var22, templ_7745c5c3_Err = templ.JoinStringErrs(placeholder)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 195, Col: 29}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var22))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 35)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if value, ok := props["value"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 36)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var23 string
				templ_7745c5c3_Var23, templ_7745c5c3_Err = templ.JoinStringErrs(value)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 198, Col: 17}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var23))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 37)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if required, ok := props["required"].(bool); ok && required {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 38)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 39)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		case "textarea":
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 40)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 41)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var24 string
				templ_7745c5c3_Var24, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 208, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var24))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 42)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if rows, ok := props["rows"].(int); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 43)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var25 string
				templ_7745c5c3_Var25, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprint(rows))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 211, Col: 27}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var25))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 44)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if cols, ok := props["cols"].(int); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 45)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var26 string
				templ_7745c5c3_Var26, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprint(cols))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 214, Col: 27}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var26))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 46)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if placeholder, ok := props["placeholder"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 47)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var27 string
				templ_7745c5c3_Var27, templ_7745c5c3_Err = templ.JoinStringErrs(placeholder)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 217, Col: 29}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var27))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 48)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 49)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if value, ok := props["value"].(string); ok {
				var templ_7745c5c3_Var28 string
				templ_7745c5c3_Var28, templ_7745c5c3_Err = templ.JoinStringErrs(value)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 222, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var28))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 50)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		case "checkbox":
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 51)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var29 string
				templ_7745c5c3_Var29, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 230, Col: 12}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var29))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 52)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var30 string
				templ_7745c5c3_Var30, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("logToConsole('%s ' + (this.checked ? 'checked' : 'unchecked'))", id))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 231, Col: 106}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var30))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 53)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if name, ok := props["name"].(string); ok {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 54)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var31 string
					templ_7745c5c3_Var31, templ_7745c5c3_Err = templ.JoinStringErrs(name)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 233, Col: 17}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var31))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 55)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				if checked, ok := props["checked"].(bool); ok && checked {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 56)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				if required, ok := props["required"].(bool); ok && required {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 57)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 58)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if label, ok := props["label"].(string); ok {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 59)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var32 string
					templ_7745c5c3_Var32, templ_7745c5c3_Err = templ.JoinStringErrs(id)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 244, Col: 45}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var32))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 60)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var33 string
					templ_7745c5c3_Var33, templ_7745c5c3_Err = templ.JoinStringErrs(label)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 244, Col: 55}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var33))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 61)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 62)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		case "list":
			if typ, ok := props["type"].(string); ok {
				if typ == "ul" {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 63)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if items, ok := props["items"].([]interface{}); ok {
						for _, item := range items {
							templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 64)
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							switch i := item.(type) {
							case string:
								var templ_7745c5c3_Var34 string
								templ_7745c5c3_Var34, templ_7745c5c3_Err = templ.JoinStringErrs(i)
								if templ_7745c5c3_Err != nil {
									return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 257, Col: 12}
								}
								_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var34))
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
							case map[string]interface{}:
								for typ, props := range i {
									templ_7745c5c3_Err = renderComponent(typ, props.(map[string]interface{})).Render(ctx, templ_7745c5c3_Buffer)
									if templ_7745c5c3_Err != nil {
										return templ_7745c5c3_Err
									}
								}
							}
							templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 65)
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						}
					}
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 66)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else if typ == "ol" {
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 67)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if items, ok := props["items"].([]interface{}); ok {
						for _, item := range items {
							templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 68)
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							switch i := item.(type) {
							case string:
								var templ_7745c5c3_Var35 string
								templ_7745c5c3_Var35, templ_7745c5c3_Err = templ.JoinStringErrs(i)
								if templ_7745c5c3_Err != nil {
									return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 274, Col: 12}
								}
								_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var35))
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
							case map[string]interface{}:
								for typ, props := range i {
									templ_7745c5c3_Err = renderComponent(typ, props.(map[string]interface{})).Render(ctx, templ_7745c5c3_Buffer)
									if templ_7745c5c3_Err != nil {
										return templ_7745c5c3_Err
									}
								}
							}
							templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 69)
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						}
					}
					templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 70)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			}
		case "form":
			if id, ok := props["id"].(string); ok {
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 71)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var36 string
				templ_7745c5c3_Var36, templ_7745c5c3_Err = templ.JoinStringErrs(id)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 289, Col: 11}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var36))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 72)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var37 string
				templ_7745c5c3_Var37, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("event.preventDefault(); logFormSubmit('%s', this)", id))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/ui-server/templates.templ`, Line: 290, Col: 92}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var37))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 73)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if components, ok := props["components"].([]interface{}); ok {
					for _, comp := range components {
						if c, ok := comp.(map[string]interface{}); ok {
							for typ, props := range c {
								templ_7745c5c3_Err = renderComponent(typ, props.(map[string]interface{})).Render(ctx, templ_7745c5c3_Buffer)
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
							}
						}
					}
				}
				templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, 74)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		}
		return templ_7745c5c3_Err
	})
}

func yamlString(def UIDefinition) string {
	yamlBytes, err := yaml.Marshal(def)
	if err != nil {
		return fmt.Sprintf("Error marshaling YAML: %v", err)
	}
	return string(yamlBytes)
}

var _ = templruntime.GeneratedTemplate
