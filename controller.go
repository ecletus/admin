package admin

import (
	"crypto/md5"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/aghape/aghape/serializer"
	"github.com/aghape/responder"
	"github.com/moisespsena-go/valuesmap"
)

// Controller admin controller
type Controller struct {
	*Admin
	action *Action
}

// HTTPUnprocessableEntity error status code
const HTTPUnprocessableEntity = 422

// Dashboard render dashboard page
func (ac *Controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *Controller) LoadIndexData(context *Context) interface{} {
	var result interface{}
	if context.ValidateLayout() {
		var err error
		result, err = context.FindMany()
		context.AddError(err)
	}
	return result
}

// Index render index page
func (ac *Controller) Index(context *Context) {
	context.Type = INDEX
	context.DefaulLayout()
	responder.With("html", func() {
		var result interface{}
		if context.LoadDisplayOrError() {
			result = ac.LoadIndexData(context)
		}
		context.Execute("", result)
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result := ac.LoadIndexData(context)
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

// New render new page
func (ac *Controller) New(context *Context) {
	context.Type = NEW
	context.Execute("", context.Resource.NewStruct(context.Context.Site))
}

// Create create data
func (ac *Controller) Create(context *Context) {
	context.Type = NEW
	res := context.Resource
	result := res.NewStruct(context.Context.Site)
	if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
		context.AddError(res.Save(result, context.Context))
	}

	if context.HasError() {
		responder.With("html", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Execute("", result)
		}).With([]string{"json", "xml"}, func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		context.Type = SHOW
		context.DefaulLayout()
		responder.With("html", func() {
			context.Flash(string(context.t(I18NGROUP+".form.successfully_created", "{{.}} was successfully created", res)), "success")
			http.Redirect(context.Writer, context.Request, context.URLFor(result, res), http.StatusFound)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(result)
		}).Respond(context.Request)
	}
}

func (ac *Controller) renderSingleton(context *Context) (interface{}, bool, error) {
	var result interface{}
	var err error
	res := context.Resource

	if res.Config.Singleton {
		result = res.NewStruct(context.Context.Site)
		if err = res.FindMany(result, res.ApplyDefaultFilters(context.Context)); err == aorm.ErrRecordNotFound {
			context.Type = NEW
			context.Execute("", result)
			return nil, true, nil
		}
	} else {
		result, err = context.FindOne()
	}
	return result, false, err
}

func (ac *Controller) LoadShowData(context *Context) (result interface{}, rendered bool) {
	var err error
	result, rendered, err = ac.renderSingleton(context)
	if rendered {
		return
	}
	context.AddError(err)
	return
}

// Show render show page
func (ac *Controller) Show(context *Context) {
	context.Type = SHOW
	context.DefaulLayout()
	responder.With("html", func() {
		if context.LoadDisplayOrError() {
			result, rendered := ac.LoadShowData(context)
			if !rendered {
				context.Execute("", result)
			}
		}
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result, _ := ac.LoadShowData(context)
			context.Encode(result)
		}
	}).Respond(context.Request)
}

// Edit render edit page
func (ac *Controller) Edit(context *Context) {
	context.Type = EDIT
	context.DefaulLayout()
	responder.With("html", func() {
		if context.LoadDisplayOrError() {
			result, rendered := ac.LoadShowData(context)
			if !rendered {
				context.Execute("", result)
			}
		}
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result, _ := ac.LoadShowData(context)
			context.Encode(result)
		}
	}).Respond(context.Request)
}

// Update update data
func (ac *Controller) Update(context *Context) {
	context.Type = EDIT
	context.DefaulLayout()
	if !context.ValidateLayoutOrError() {
		return
	}
	var result interface{}
	res := context.Resource

	if !context.LoadDisplayOrError() {
		return
	}

	result, _ = ac.LoadShowData(context)

	if !context.HasError() {
		decerror := res.Decode(context.Context, result)
		if context.AddError(decerror); !context.HasError() {
			context.AddError(res.Save(result, context.Context))
		}
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("", result)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		context.Type = SHOW
		context.DefaulLayout()
		responder.With("html", func() {
			context.Flash(string(context.t(I18NGROUP+".form.successfully_updated", "{{.}} was successfully updated", res)), "success")
			context.Execute("", result)
		}).With([]string{"json", "xml"}, func() {
			if context.Request.FormValue("qorInlineEdit") != "" {
				rresult := reflect.ValueOf(result)
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
						} else if gsf, ok := result.(serializer.SerializableField); ok {
							if value, ok := gsf.GetSerializableField(key); ok {
								newResult[key] = value
							}
						}
					}
				}
				newResult = valuesmap.ParseMap(newResult)
				context.Encode(newResult)
				return
			}
			context.Encode(result)
		}).Respond(context.Request)
	}
}

// Delete delete data
func (ac *Controller) Delete(context *Context) {
	res := context.Resource
	status := http.StatusOK

	if context.AddError(res.Delete(res.NewStruct(context.Context.Site), res.ApplyDefaultFilters(context.Context))); context.HasError() {
		context.Flash(string(context.t(I18NGROUP+".form.failed_to_delete", "Failed to delete {{.}}", res)), "error")
		status = http.StatusNotFound
	}

	responder.With("html", func() {
		http.Redirect(context.Writer, context.Request, res.GetContextIndexURI(context.Context), http.StatusFound)
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
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
				message := string(context.t(I18NGROUP+".actions.executed_successfully", "Action {{.Name}}: Executed successfully", action))
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
					message := string(context.t(I18NGROUP+".actions.executed_failed", "Action {{.Name}}: Failed to execute", action))
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
	file := strings.TrimPrefix(context.Request.URL.Path, ac.Router.Prefix())

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
