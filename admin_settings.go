package admin

import (
	"encoding/json"
	"fmt"

	"github.com/ecletus/common"
	"github.com/moisespsena-go/aorm"
)

// SettingsStorageInterface settings storage interface
type SettingsStorageInterface interface {
	Get(key string, value interface{}, context *Context) error
	Save(key string, value interface{}, res *Resource, user common.User, context *Context) error
}

// AdminSetting admin settings
type AdminSetting struct {
	aorm.Model
	Key      string
	Resource string
	UserID   string
	Value    string `gorm:"type:text"`
}

type settings struct{}

// Get load admin settings
func (settings) Get(key string, value interface{}, context *Context) error {
	var (
		settings  = []AdminSetting{}
		tx        = context.Site.GetSystemDB().DB.New()
		resParams = ""
		userID    = ""
	)
	sqlCondition := fmt.Sprintf("%v = ? AND (resource = ? OR resource = ?) AND (user_id = ? OR user_id = ?)", aorm.QuotePath(tx.Dialect(), "key"))

	if context.Resource != nil {
		resParams = context.Resource.ToParam()
	}

	if context.CurrentUser() != nil {
		userID = ""
	}

	tx.Where(sqlCondition, key, resParams, "", userID, "").Order("user_id DESC, resource DESC, id DESC").Find(&settings)

	for _, setting := range settings {
		if err := json.Unmarshal([]byte(setting.Value), value); err != nil {
			return err
		}
	}

	return nil
}

// Save save admin settings
func (settings) Save(key string, value interface{}, res *Resource, user common.User, context *Context) error {
	var (
		tx          = context.DB().New()
		result, err = json.Marshal(value)
		resParams   = ""
		userID      = ""
	)

	if err != nil {
		return err
	}

	if res != nil {
		resParams = res.ToParam()
	}

	if user != nil {
		userID = ""
	}

	err = tx.Where(AdminSetting{
		Key:      key,
		UserID:   userID,
		Resource: resParams,
	}).Assign(AdminSetting{Value: string(result)}).FirstOrCreate(&AdminSetting{}).Error

	return err
}

func (this *Admin) Settings() SettingsStorageInterface {
	return &this.settings
}
