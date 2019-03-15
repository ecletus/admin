package admin

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/moisespsena-go/field-access"

	"github.com/ecletus/roles"

	"github.com/ecletus/core"
	"github.com/ecletus/responder"
	"github.com/moisespsena-go/valuesmap"
)

// Controller admin controller
type Controller struct {
	*Admin
	action     *Action
	controller interface{}
}

// HTTPUnprocessableEntity error status code
const HTTPUnprocessableEntity = 422

// Dashboard render dashboard page
func (ac *Controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *Controller) LoadIndexData(context *Context) interface{} {
	return ac.controller.(ControllerIndex).Index(context)
}

// Index render index page
func (ac *Controller) Index(context *Context) {
	context.Type = INDEX
	context.DefaulLayout()
	defer context.LogErrors()
	responder.With("html", func() {
		var result interface{}
		if context.LoadDisplayOrError() {
			result = ac.LoadIndexData(context)
		}
		context.Execute("", result)
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result := ac.LoadIndexData(context)
			context.Api = true
			context.Encode(result)
		}
	}).Respond(context.Request)
}

// SearchCenter render search center page
func (ac *Controller) SearchCenter(context *Context) {
	type Result struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}
	var searchResults []Result

	for _, res := range context.GetSearchableResources() {
		var (
			resourceName = context.Request.URL.Query().Get("resource_name")
			ctx          = context.clone().setResource(res)
			searchResult = Result{Context: ctx, Resource: res}
		)

		if resourceName == "" || res.ToParam() == resourceName {
			searchResult.Results, _ = ctx.FindMany()
		}
		searchResults = append(searchResults, searchResult)
	}
	context.Execute("search_center", searchResults)
}

// Search render search page
func (ac *Controller) Search(context *Context) {
	type Result struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}

	var searchResult = Result{Context: context, Resource: context.Resource}
	searchResult.Results, _ = context.FindMany()

	context.Execute("search_center", []Result{searchResult})
}

// New render new page
func (ac *Controller) New(context *Context) {
	context.Type = NEW
	context.Execute("", ac.controller.(ControllerCreator).New(context))
}

// Create create data
func (ac *Controller) Create(context *Context) {
	context.Type = NEW
	res := context.Resource
	recorde := ac.controller.(ControllerCreator).New(context)

	if !context.HasError() {
		if context.AddError(res.Decode(context.Context, recorde)); !context.HasError() {
			ac.controller.(ControllerCreator).Create(context, recorde)
		}
	}

	if context.HasError() {
		responder.With("html", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Execute("", recorde)
		}).With([]string{"json", "xml"}, func() {
			context.Api = true
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		defer context.LogErrors()
		context.Type = SHOW
		context.DefaulLayout()
		context.Flash(string(context.tt(I18NGROUP+".form.successfully_created",
			NewResourceRecorde(context, res, recorde),
			"{{.}} was successfully created")), "success")
		responder.With("html", func() {
			url := context.RedirectTo
			if url == "" {
				if context.Request.URL.Query().Get("continue_editing") != "" {
					url = res.GetContextURI(context.Context, res.GetKey(recorde)) + P_OBJ_UPDATE_FORM
				} else if context.Request.URL.Query().Get("continue_editing_url") != "" {
					url = res.GetContextURI(context.Context, res.GetKey(recorde)) + P_OBJ_UPDATE_FORM
					context.Writer.Header().Set("X-Location", url)
					context.Writer.WriteHeader(http.StatusNoContent)
					return
				} else {
					url = res.GetContextIndexURI(context.Context)
				}
			}
			http.Redirect(context.Writer, context.Request, url, http.StatusFound)
		}).With([]string{"json", "xml"}, func() {
			context.Api = true
			if context.Request.URL.Query().Get("continue_editing") != "" {
				url := res.GetContextURI(context.Context, res.GetKey(recorde)) + P_OBJ_UPDATE_FORM
				context.Encode(map[string]interface{}{"HTTPRedirectTo": url})
				return
			}
			context.Encode(recorde)
		}).Respond(context.Request)
	}
}

func (ac *Controller) LoadShowData(context *Context) (result interface{}) {
	res := context.Resource

	if res.Config.Singleton {
		if reader, ok := ac.controller.(ControllerReader); ok {
			result = reader.Read(context)
		} else if creator, ok := ac.controller.(ControllerCreator); ok {
			result = creator.New(context)
		} else {
			result = res.NewStruct(context.Site)
		}
		if result == nil {
			if core.HasPermission(res, roles.Create, context.Context) {
				if creator, ok := ac.controller.(ControllerCreator); ok {
					result = creator.New(context)
				} else {
					result = res.NewStruct(context.Site)
				}
				context.Type = NEW
			}
		}
	} else {
		context.SetDB(context.DB.Unscoped())
		result = ac.controller.(ControllerReader).Read(context)
	}
	return
}

// Show render show page
func (ac *Controller) showOrEdit(context *Context) {
	context.DefaulLayout()
	responder.With("html", func() {
		if context.LoadDisplayOrError() {
			recorde := ac.LoadShowData(context)
			if !context.HasError() {
				if recorde == nil {
					context.NotFound = true
					http.NotFound(context.Writer, context.Request)
					return
				}
				if !context.Type.Has(DELETED) && context.Resource.IsSoftDeleted(recorde) {
					http.Redirect(context.Writer, context.Request, context.Resource.GetContextIndexURI(context.Context), http.StatusSeeOther)
					return
				}
			}
			context.Execute("", recorde)
		}
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			recorde := ac.LoadShowData(context)
			if !context.HasError() {
				if recorde == nil {
					context.NotFound = true
					http.NotFound(context.Writer, context.Request)
					return
				}
			} else {
				context.Writer.WriteHeader(http.StatusBadGateway)
				context.Writer.Write([]byte(context.Error()))
			}
			context.Encode(recorde)
		}
	}).Respond(context.Request)
}

// Show render show page
func (ac *Controller) Show(context *Context) {
	context.Type = SHOW
	if HasDeletedUrlQuery(context.Request.URL.Query()) {
		context.Type |= DELETED
	}
	ac.showOrEdit(context)
}

// Edit render edit page
func (ac *Controller) Edit(context *Context) {
	context.Type = EDIT
	ac.showOrEdit(context)
}

// Update update data
func (ac *Controller) Update(context *Context) {
	context.Type = EDIT
	context.DefaulLayout()
	if !context.ValidateLayoutOrError() {
		return
	}
	var recorde interface{}
	res := context.Resource

	if !context.LoadDisplayOrError() {
		return
	}

	recorde = ac.LoadShowData(context)

	if !context.HasError() {
		decerror := res.Decode(context.Context, recorde)
		if context.AddError(decerror); !context.HasError() {
			ac.controller.(ControllerUpdater).Update(context, recorde)
		}
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("", recorde)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		defer context.LogErrors()
		context.Type = SHOW
		context.DefaulLayout()
		responder.With("html", func() {
			context.Flash(string(context.tt(I18NGROUP+".form.successfully_updated", NewResourceRecorde(context, res, recorde),
				"{{.}} was successfully updated")), "success")
			url := context.RedirectTo
			if url == "" {
				if res.Config.Singleton {
					url = res.GetContextIndexURI(context.Context)
				} else {
					url = res.GetContextURI(context.Context, res.GetKey(recorde))
				}
				url += P_OBJ_UPDATE_FORM
			}
			if context.Request.URL.Query().Get("continue_editing") != "" {
				http.Redirect(context.Writer, context.Request, url, http.StatusFound)
				return
			} else if context.Request.URL.Query().Get("continue_editing_url") != "" {
				context.Writer.Header().Set("X-Location", url)
				context.Writer.WriteHeader(http.StatusNoContent)
				return
			}
			http.Redirect(context.Writer, context.Request, url, http.StatusFound)
		}).With([]string{"json", "xml"}, func() {
			if context.Request.FormValue("qorInlineEdit") != "" {
				rresult := reflect.ValueOf(recorde)
				for rresult.Kind() == reflect.Ptr {
					rresult = rresult.Elem()
				}
				newResult := make(map[string]interface{})

				for key, _ := range context.Request.Form {
					if strings.HasPrefix(key, "QorResource.") {
						key = strings.TrimPrefix(key, "QorResource.")
						f := rresult.FieldByName(key)
						if f.IsValid() {
							newResult[key] = f.Interface()
						} else if v, ok := field_access.Get(recorde, key); ok {
							newResult[key] = v
						}
					}
				}
				newResult = valuesmap.ParseMap(newResult)
				context.Encode(newResult)
				return
			}
			context.Encode(recorde)
		}).Respond(context.Request)
	}
}

// Delete delete data
func (ac *Controller) Delete(context *Context) {
	res := context.Resource
	status := http.StatusOK
	recorde := ac.controller.(ControllerReader).Read(context)
	var msg string

	if recorde == nil && !context.HasError() {
		status = http.StatusNotFound
	} else {
		if context.HasError() {
			msg = string(context.tt(I18NGROUP+".form.failed_to_delete",
				NewResourceRecorde(context, res, recorde),
				"Failed to delete {{.}}"))
			status = http.StatusBadRequest
		} else {
			ac.controller.(ControllerDeleter).Delete(context, recorde)
			if context.HasError() {
				status = http.StatusBadRequest
				msg = string(context.tt(I18NGROUP+".form.failed_to_delete_error",
					map[string]interface{}{
						"Value": NewResourceRecorde(context, res, recorde),
						"Error": context.Errors,
					},
					"Failed to delete {{.Value}}: {{.Error}}"))
			} else {
				msg = string(context.tt(I18NGROUP+".form.successfully_deleted", NewResourceRecorde(context, res, recorde),
					"{{.}} was successfully deleted"))
			}
		}
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			uri := res.GetContextIndexURI(context.Context)
			http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
			return
		} else {
			context.Writer.WriteHeader(status)
			context.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		}
		context.Layout = "OK"
		context.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := res.GetContextIndexURI(context.Context)
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}

// BulkDelete delete many recordes
func (ac *Controller) BulkDelete(context *Context) {
	var (
		res      = context.Resource
		status   = http.StatusOK
		recordes []interface{}
		keySep   = context.Request.URL.Query().Get("key_sep")
		keys     []string
		msg      string
	)

	if context.Request.ContentLength > 0 {
		if context.Request.Header.Get("Content-Type") == "application/json" {
			err := json.NewDecoder(context.Request.Body).Decode(&keys)
			if err != nil {
				status = http.StatusBadRequest
				context.AddError(err)
			}
		} else {
			keys = context.Request.Form["primary_values[]"]
		}
	} else {
		if keySep == "" {
			keySep = ":"
		}
		keys = strings.Split(context.Request.URL.Query().Get("key"), keySep)
	}

	if !context.HasError() {
		for _, key := range keys {
			if key == "" {
				continue
			}

			context.ResourceID = key
			recorde := ac.controller.(ControllerReader).Read(context)
			if !context.HasError() {
				recordes = append(recordes, recorde)
			}
		}
	}

	context.ResourceID = ""

	if len(recordes) == 0 && !context.HasError() {
		status = http.StatusNotFound
	} else {
		if context.HasError() {
			msg = string(context.tt(I18NGROUP+".form.failed_to_delete",
				NewResourceRecorde(context, res, recordes[len(recordes)-1]),
				"Failed to delete {{.}}"))
			status = http.StatusBadRequest
		} else {
			ac.controller.(ControllerBulkDeleter).DeleteBulk(context, recordes...)
		}
	}

	if msg == "" && !context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.successfully_bulk_deleted", NewResourceRecorde(context, res, recordes...),
			"{{.}} was successfully deleted"))
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			uri := res.GetContextIndexURI(context.Context, context.Context.ParentResourceID...)
			http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
		} else {
			context.Writer.WriteHeader(status)
			context.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		}
		if msg == "" && context.HasError() {
			msg = context.Error()
		}
		context.Layout = "OK"
		context.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := res.GetContextIndexURI(context.Context, context.Context.ParentResourceID...)
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}

// Delete delete data
func (ac *Controller) DeletedIndex(context *Context) {
	context.Type = INDEX | DELETED
	context.DefaulLayout()
	ctrl := ac.controller.(ControllerRestorer)
	defer context.LogErrors()
	responder.With("html", func() {
		var result interface{}
		if context.LoadDisplayOrError() {
			result = ctrl.DeletedIndex(context)
		}
		context.Execute(INDEX.S(), result)
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result := ctrl.DeletedIndex(context)
			context.Api = true
			context.Encode(result)
		}
	}).Respond(context.Request)
}

// BulkDelete delete many recordes
func (ac *Controller) Restore(context *Context) {
	var (
		res    = context.Resource
		status = http.StatusOK
		keySep = context.Request.URL.Query().Get("key_sep")
		keys   []string
		msg    string
	)

	if context.Request.ContentLength > 0 {
		if context.Request.Header.Get("Content-Type") == "application/json" {
			err := json.NewDecoder(context.Request.Body).Decode(&keys)
			if err != nil {
				status = http.StatusBadRequest
				context.AddError(err)
			}
		} else {
			keys = context.Request.Form["primary_values[]"]
		}
	} else {
		if keySep == "" {
			keySep = ":"
		}
		keys = strings.Split(context.Request.URL.Query().Get("key"), keySep)
	}

	if len(keys) == 0 {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		return
	}

	context.Type |= DELETED

	context.ResourceID = ""

	if context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.failed_to_restore",
			NewResourceRecorde(context, res),
			"Failed to restore {{.}}"))
		status = http.StatusBadRequest
	} else {
		ac.controller.(ControllerRestorer).Restore(context, keys...)
	}

	if msg == "" && !context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.successfully_restored", NewResourceRecorde(context, res),
			"{{.}} was successfully restored"))
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			url := res.GetContextIndexURI(context.Context)

			if context.Request.URL.Query().Get("continue_editing") != "" {
				http.Redirect(context.Writer, context.Request, url+"/"+keys[0], http.StatusFound)
				return
			} else if context.Request.URL.Query().Get("continue_editing_url") != "" {
				context.Writer.Header().Set("X-Continue-Editing-Url", url+"/"+keys[0])
				context.Writer.WriteHeader(http.StatusNoContent)
				return
			}

			http.Redirect(context.Writer, context.Request, url, http.StatusFound)
		} else {
			context.Writer.WriteHeader(status)
			context.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		}
		if msg == "" && context.HasError() {
			msg = context.Error()
		}
		context.Layout = "OK"
		context.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := res.GetContextIndexURI(context.Context, context.Context.ParentResourceID...)
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}

// Action handle action related requests
func (ac *Controller) Action(context *Context) {
	var action = ac.action

	if action.Available != nil {
		if !action.Available(context) {
			context.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	defer context.LogErrors()

	if context.Request.Method == "GET" {
		context.Execute("action", action)
	} else {
		var actionArgument = ActionArgument{
			PrimaryValues: context.Request.Form["primary_values[]"],
			Context:       context,
		}

		if primaryValue := context.URLParam(context.Resource.ParamIDName()); primaryValue != "" {
			actionArgument.PrimaryValues = append(actionArgument.PrimaryValues, primaryValue)
		}

		if action.Resource != nil {
			result := action.Resource.NewStruct(context.Context.Site)
			action.Resource.Decode(context.Context, result)
			actionArgument.Argument = result
		}

		err := action.Handler(&actionArgument)

		if !actionArgument.SkipDefaultResponse {
			if err == nil {
				message := string(context.tt(I18NGROUP+".actions.executed_successfully", action, "Action {{.Name}}: Executed successfully"))
				responder.With("html", func() {
					context.Flash(message, "success")
					http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
				}).With([]string{"json"}, func() {
					context.Layout = "OK"
					context.Encode(map[string]string{"message": message, "status": "ok"})
				}).Respond(context.Request)
			} else {
				context.Writer.WriteHeader(HTTPUnprocessableEntity)
				responder.With("html", func() {
					context.AddError(err)
					context.Execute("action", action)
				}).With([]string{"json", "xml"}, func() {
					context.Layout = "OK"
					message := string(context.tt(I18NGROUP+".actions.executed_failed", action, "Action {{.Name}}: Failed to execute"))
					context.Encode(map[string]string{"error": message, "status": "error"})
				}).Respond(context.Request)
			}
		}
	}
}

var (
	cacheSince = time.Now().Format(http.TimeFormat)
)

// Asset handle asset requests
func (ac *Controller) Asset(context *Context) {
	var done bool

	context.SetupConfig().IfProd(func() error {
		if context.Request.Header.Get("If-Modified-Since") == cacheSince {
			context.Writer.WriteHeader(http.StatusNotModified)
			done = true
			return nil
		}
		context.Writer.Header().Set("Last-Modified", cacheSince)
		return nil
	})

	if done {
		return
	}
	file := strings.TrimPrefix(context.Request.URL.Path, ac.Config.MountPath)

	if asset, err := context.Asset(file); err == nil {
		var etag string
		context.SetupConfig().IfProd(func() (err error) {
			etag = fmt.Sprintf("%x", md5.Sum(asset.GetData()))
			if context.Request.Header.Get("If-None-Match") == etag {
				context.Writer.WriteHeader(http.StatusNotModified)
				done = true
				return
			}
			return
		})

		if done {
			return
		}

		if ctype := mime.TypeByExtension(filepath.Ext(file)); ctype != "" {
			context.Writer.Header().Set("Content-Type", ctype)
		}

		context.SetupConfig().IfProd(func() error {
			context.Writer.Header().Set("Cache-control", "private, must-revalidate, max-age=300")
			context.Writer.Header().Set("ETag", etag)
			return nil
		})
		context.Writer.Write(asset.GetData())
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
