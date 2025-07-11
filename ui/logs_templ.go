// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.898
package ui

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"

	i "github.com/callsamu/templicons"
	"gotail/models"
	"gotail/ui/components"
)

func onOpenDrawer(id string) templ.ComponentScript {
	return templ.ComponentScript{
		Name: `__templ_onOpenDrawer_2ce8`,
		Function: `function __templ_onOpenDrawer_2ce8(id){document.getElementById(id).classList.toggle("lg:w-96");
  document.getElementById(id).classList.toggle("w-72");
  document.getElementById("body").classList.toggle("overflow-hidden");

  const backdrop = document.getElementById(id + "-backdrop");
  backdrop.classList.toggle("opacity-100");

  if (backdrop.classList.contains("opacity-100")) {
    backdrop.classList.remove("pointer-events-none");
    backdrop.classList.add("pointer-events-auto");
  } else {
    backdrop.classList.remove("pointer-events-auto");
    backdrop.classList.add("pointer-events-none");
  }
}`,
		Call:       templ.SafeScript(`__templ_onOpenDrawer_2ce8`, id),
		CallInline: templ.SafeScriptInline(`__templ_onOpenDrawer_2ce8`, id),
	}
}

func firstN(n int, attrs map[string]any) map[string]any {
	result := make(map[string]any)
	count := 0
	for k, v := range attrs {
		if count >= n {
			break
		}
		result[k] = v
		count++
	}
	return result
}

func LogsView(data struct {
	Logs       []models.LogEntry
	Page       int
	Limit      int
	Total      int
	Severity   string
	AttrKeys   []string
	AttrValue  string
	AttrKey    string
	CurrentUrl string
	TotalLogs  int
	Services   []string
	Service    string
}) templ.Component {
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
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\" class=\"w-full h-full bg-gray-50/40 text-gray-900\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>GoTail - Logs</title><script src=\"https://cdn.tailwindcss.com\"></script><style>\r\n        html, body {\r\n          height: 100%;\r\n          margin: 0;\r\n          padding: 0;\r\n        }\r\n      </style></head><body id=\"body\" class=\"w-full h-full\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.Sidebar(struct{ CurrentUrl string }{CurrentUrl: data.CurrentUrl}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, item := range data.Logs {
			templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<div class=\"space-y-8\"><div class=\"space-y-2\"><h2 class=\"text-2xl font-semibold\">Log Entry Details</h2><p class=\"text-sm text-gray-500\">Detailed information for log entry ID ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var3 string
				templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(item.ID)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 82, Col: 62}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</p></div><div class=\"grid grid-cols-2 gap-x-4\"><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Severity</h3>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = components.Severity(struct{ Severity string }{Severity: item.SeverityText}).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</div><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Timestamp</h3><p class=\"text-sm\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(item.Timestamp.Format("2006-01-02 15:04:05"))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 99, Col: 63}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</p></div></div><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Message</h3><div class=\"p-2 bg-gray-100 rounded-lg w-full\"><p class=\"text-sm\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(item.Body)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 110, Col: 28}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "</p></div></div><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Service</h3><p class=\"text-sm\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if item.ServiceName != nil && *item.ServiceName != "" {
					var templ_7745c5c3_Var6 string
					templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(*item.ServiceName)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 121, Col: 36}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "<span class=\"text-gray-400\">No service name</span>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "</p></div><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Host</h3><p class=\"text-sm\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if item.HostName != nil && *item.HostName != "" {
					var templ_7745c5c3_Var7 string
					templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(*item.HostName)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 134, Col: 33}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "<span class=\"text-gray-400\">No host name</span>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</p></div><div class=\"text-start space-y-2\"><h3 class=\"text-sm font-medium text-gray-500\">Attributes</h3><div class=\"space-y-2\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				for k, v := range item.Attributes {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "<div class=\"p-2 bg-gray-100 rounded-lg w-full flex items-center justify-between space-x-4\"><span class=\"text-sm font-medium\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var8 string
					templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(k)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 148, Col: 56}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "</span> <span class=\"text-sm text-end\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var9 string
					templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%v", v))
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 149, Col: 72}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "</span></div>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "</div></div></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				return nil
			})
			templ_7745c5c3_Err = components.Drawer(struct{ ID string }{ID: fmt.Sprintf("log-%d", item.ID)}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "<div class=\"lg:ml-64 px-2 py-6 lg:p-8 space-y-6\"><div class=\"flex items-start justify-between space-x-8\"><div class=\"space-y-2\"><h1 class=\"text-2xl lg:text-3xl font-bold\">Log Entries</h1><p class=\"text-sm lg:text-md text-gray-500\">Browse and filter log entries from your application.</p></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.MobileSidebar(struct{ CurrentUrl string }{CurrentUrl: data.CurrentUrl}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, "</div><div class=\"w-full p-6 rounded-lg shadow-sm border space-y-4 bg-white\"><div class=\"flex items-center space-x-2\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = i.Icon("mdi:search", i.Params().SetDimensions(24, 24)).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, "<p class=\"text-2xl font-semibold\">Filters</p></div><form method=\"GET\" class=\"grid lg:grid-cols-2 lg:gap-x-4 gap-y-4 lg:gap-y-8 lg:items-end\"><!-- Severity filter --><div class=\"space-y-2\"><label for=\"severity\" class=\"block text-sm font-medium\">Severity Level</label> <select name=\"severity\" class=\"w-full border p-2 rounded-lg\"><option value=\"\">All levels</option> <option value=\"INFO\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Severity == "INFO" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 19, ">INFO</option> <option value=\"WARNING\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Severity == "WARNING" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 20, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 21, ">WARNING</option> <option value=\"ERROR\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Severity == "ERROR" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 22, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 23, ">ERROR</option> <option value=\"DEBUG\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Severity == "DEBUG" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 24, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 25, ">DEBUG</option> <option value=\"FATAL\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Severity == "FATAL" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 26, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 27, ">FATAL</option></select></div><!-- Service name dropdown --><div class=\"space-y-2\"><label for=\"service\" class=\"block text-sm font-medium\">Service Name</label> <select name=\"service\" class=\"w-full border p-2 rounded-lg\"><option value=\"\">Any service</option> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, service := range data.Services {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 28, "<option value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var10 string
			templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(service)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 208, Col: 34}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 29, "\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if data.Service == service {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 30, " selected")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 31, ">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var11 string
			templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(service)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 211, Col: 28}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 32, "</option>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 33, "</select></div><!-- Attribute key dropdown --><div class=\"space-y-2\"><label for=\"attr_key\" class=\"block text-sm font-medium\">Attribute Key</label> <select name=\"attr_key\" class=\"w-full border p-2 rounded-lg\"><option value=\"\">Any attribute</option> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, key := range data.AttrKeys {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 34, "<option value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var12 string
			templ_7745c5c3_Var12, templ_7745c5c3_Err = templ.JoinStringErrs(key)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 231, Col: 30}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var12))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 35, "\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if data.AttrKey == key {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 36, " selected")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 37, ">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var13 string
			templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(key)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 234, Col: 24}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 38, "</option>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 39, "</select></div><!-- Attribute value input --><div class=\"space-y-2\"><label for=\"attr_value\" class=\"block text-sm font-medium\">Attribute Value</label> <input type=\"text\" name=\"attr_value\" placeholder=\"Search by value...\" class=\"w-full border p-2 rounded-lg\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var14 string
		templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs(data.AttrValue)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 250, Col: 37}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 40, "\"></div><input type=\"hidden\" name=\"page\" value=\"1\"> <button type=\"submit\" class=\"bg-[#0f172a] border border-[#0f172a] text-[#f8fafc] h-[42px] px-4 py-2 rounded-lg\">Search</button> <button type=\"reset\" class=\"bg-gray-300 border border-gray-300 text-gray-900 px-4 py-2 rounded-lg h-[42px]\" onclick=\"window.location.href='/'\">Reset</button></form></div><div class=\"w-full p-6 rounded-lg shadow-sm border space-y-4 bg-white\"><h1 class=\"text-xl font-semibold\">Log Entries (")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var15 string
		templ_7745c5c3_Var15, templ_7745c5c3_Err = templ.JoinStringErrs(data.TotalLogs)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 274, Col: 40}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var15))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 41, " total) </h1></div><div class=\"flex justify-end\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.Pagination(struct {
			Page       int
			Level      string
			AttrKey    string
			AttrValue  string
			TotalPages int
		}{
			Page:       data.Page,
			Level:      data.Severity,
			AttrKey:    data.AttrKey,
			AttrValue:  data.AttrValue,
			TotalPages: (data.Total + data.Limit - 1) / data.Limit,
		}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 42, "</div><div class=\"block lg:hidden space-y-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, item := range data.Logs {
			templ_7745c5c3_Err = templ.RenderScriptItems(ctx, templ_7745c5c3_Buffer, onOpenDrawer(fmt.Sprintf("log-%d", item.ID)))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 43, "<div class=\"rounded-lg border bg-white shadow p-4 space-y-4 cursor-pointer\" onClick=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var16 templ.ComponentScript = onOpenDrawer(fmt.Sprintf("log-%d", item.ID))
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var16.Call)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 44, "\"><div class=\"flex w-full justify-between items-center\"><p class=\"text-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var17 string
			templ_7745c5c3_Var17, templ_7745c5c3_Err = templ.JoinStringErrs(item.Timestamp.Format("2006-01-02 15:04:05"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 301, Col: 80}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var17))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 45, "</p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = components.Severity(struct{ Severity string }{Severity: item.SeverityText}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 46, "</div><div class=\"space-y-2\"><p class=\"text-sm text-gray-500 mt-2\">Service</p><div class=\"text-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if item.ServiceName != nil && *item.ServiceName != "" {
				var templ_7745c5c3_Var18 string
				templ_7745c5c3_Var18, templ_7745c5c3_Err = templ.JoinStringErrs(*item.ServiceName)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 310, Col: 38}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var18))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 47, "<span class=\"text-gray-400\">N/A</span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 48, "</div></div><div class=\"space-y-2\"><p class=\"text-sm text-gray-500 mt-2\">Message</p><p class=\"text-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var19 string
			templ_7745c5c3_Var19, templ_7745c5c3_Err = templ.JoinStringErrs(item.Body)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 319, Col: 45}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var19))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 49, "</p></div><div class=\"space-y-2\"><p class=\"text-sm text-gray-500 mt-2\">Attributes</p><div class=\"text-sm space-y-1\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for k, v := range firstN(3, item.Attributes) {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 50, "<div><span class=\"text-gray-600\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var20 string
				templ_7745c5c3_Var20, templ_7745c5c3_Err = templ.JoinStringErrs(k)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 327, Col: 52}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var20))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 51, "</span>: ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var21 string
				templ_7745c5c3_Var21, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%v", v))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 327, Col: 83}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var21))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 52, "</div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if len(item.Attributes) > 3 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 53, "<span class=\"inline-block bg-gray-100 py-1 px-2 text-xs rounded mt-1\">+ ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var22 string
				templ_7745c5c3_Var22, templ_7745c5c3_Err = templ.JoinStringErrs(len(item.Attributes) - 3)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 332, Col: 49}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var22))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 54, " more</span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 55, "</div></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 56, "</div><table class=\"hidden lg:table w-full bg-white shadow rounded overflow-hidden\"><thead class=\"bg-gray-100 text-left text-sm font-semibold\"><tr><th class=\"p-2\">Time</th><th class=\"p-2\">Level</th><th class=\"p-2\">Service</th><th class=\"p-2 w-[500px]\">Message</th><th class=\"p-2\">Attributes</th></tr></thead> <tbody>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, item := range data.Logs {
			templ_7745c5c3_Err = templ.RenderScriptItems(ctx, templ_7745c5c3_Buffer, onOpenDrawer(fmt.Sprintf("log-%d", item.ID)))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 57, "<tr class=\"border-t cursor-pointer\" onClick=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var23 templ.ComponentScript = onOpenDrawer(fmt.Sprintf("log-%d", item.ID))
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var23.Call)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 58, "\"><td class=\"p-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var24 string
			templ_7745c5c3_Var24, templ_7745c5c3_Err = templ.JoinStringErrs(item.Timestamp.Format("2006-01-02 15:04:05"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 357, Col: 77}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var24))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 59, "</td><td class=\"p-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = components.Severity(struct{ Severity string }{Severity: item.SeverityText}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 60, "</td><td class=\"2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if item.ServiceName != nil && *item.ServiceName != "" {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 61, "<span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var25 string
				templ_7745c5c3_Var25, templ_7745c5c3_Err = templ.JoinStringErrs(*item.ServiceName)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 363, Col: 44}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var25))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 62, "</span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 63, "<span class=\"text-gray-400\">N/A</span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 64, "</td><td class=\"p-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var26 string
			templ_7745c5c3_Var26, templ_7745c5c3_Err = templ.JoinStringErrs(item.Body)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 368, Col: 42}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var26))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 65, "</td><td class=\"p-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for k, v := range firstN(3, item.Attributes) {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 66, "<div><span class=\"text-gray-600\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var27 string
				templ_7745c5c3_Var27, templ_7745c5c3_Err = templ.JoinStringErrs(k)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 372, Col: 52}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var27))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 67, "</span>: <span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var28 string
				templ_7745c5c3_Var28, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%v", v))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 373, Col: 49}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var28))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 68, "</span></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if len(item.Attributes) > 3 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 69, "<span class=\"flex rounded-lg bg-gray-100 py-1 px-2 text-xs block w-fit mt-2\">+ ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var29 string
				templ_7745c5c3_Var29, templ_7745c5c3_Err = templ.JoinStringErrs(len(item.Attributes) - 3)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `ui/logs.templ`, Line: 378, Col: 49}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var29))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 70, " more</span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 71, "</td></tr>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 72, "</tbody></table><div class=\"mt-4 flex justify-end\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.Pagination(struct {
			Page       int
			Level      string
			AttrKey    string
			AttrValue  string
			TotalPages int
		}{
			Page:       data.Page,
			Level:      data.Severity,
			AttrKey:    data.AttrKey,
			AttrValue:  data.AttrValue,
			TotalPages: (data.Total + data.Limit - 1) / data.Limit,
		}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 73, "</div></div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
