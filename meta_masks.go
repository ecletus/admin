package admin

import "fmt"

const (
	MASK_CPF             = "cpf"
	MASK_CNPJ            = "cnpj"
	MASK_CPF_OR_CNPJ     = "cpf_or_cnpj"
	MASK_DV1             = "dv1"
	MASK_DV2             = "dv2"
	MASK_MONEY           = "money"
	MASK_PHONE_BR        = "phone_br"
	MASK_PHONE_MOBILE_BR = "phone_mobile_br"
	MASK_PHONE_FIXO_BR   = "phone_fixo_br"
	MASK_ZIPCODE         = "zipcode"
	MASK_CAR_TAG_BR      = "car_tag_br"
)

var Masks map[string]*MaskConfig

func GetMask(name string) *MaskConfig {
	return Masks[name]
}

func RegisterMask(name string, config *MaskConfig) {
	if _, ok := Masks[name]; ok {
		panic(fmt.Errorf("duplicate mask %q", name))
	}
	Masks[name] = config
}

func init() {
	Masks = map[string]*MaskConfig{
		MASK_CAR_TAG_BR: {
			JsCode: `
let mask = function (value, e, $field, options) {
        // Convert to uppercase
        value = value.toUpperCase();
		if (value && e) e.currentTarget.value = value;

        // Get only valid characters
        let val = value.replace(/[^\w]/g, '');

        // Detect plate format
        let isNumeric = !isNaN(parseFloat(val[4])) && isFinite(val[4]);
        let mask = 'AAA 0U00';
        if(val.length > 4 && isNumeric) {
            mask = 'AAA-0000';
        }
        $field.mask(mask, options);
    }.bind(this),
	options = {
		translation: {
			'A': {
				pattern: /[A-Za-z]/
			},
			'U': {
				pattern: /[A-Za-z0-9]/
			},
		},
		onKeyPress: mask
	},
	val = this.val();

this.mask('AAA 0U00', options);

if (val) mask(val, null, this, options);
`,
		},
		MASK_CPF: {
			JsCode: `this.mask("000.000.000-00", {reverse: true})`,
		},
		MASK_CNPJ: {
			JsCode: `this.mask("00.000.000/0000-00", {reverse: true})`,
		},
		MASK_CPF_OR_CNPJ: {
			JsCode: `var behavior = function (val) {
    return val.replace(/\D/g, '').length <= 11 ? '000.000.000-009' : '00.000.000/0000-00'
},
options = {
    onKeyPress: function (val, e, field, options) {
        field.mask(behavior.apply({}, arguments), options);
    }
};

this.mask(behavior, options)`,
		},
		MASK_MONEY: {
			JsCode: `this.mask("#.##0,00", {reverse: true})`,
		},
		MASK_DV1: {
			JsCode: `this.mask("#0-0", {reverse: true})`,
		},
		MASK_DV2: {
			JsCode: `this.mask("#0-00", {reverse: true})`,
		},
		MASK_ZIPCODE: {
			JsCode: `this.mask("00000-000")`,
		},
		MASK_PHONE_FIXO_BR: {
			JsCode: `this.mask("(00) 0000-0000")`,
		},
		MASK_PHONE_MOBILE_BR: {
			JsCode: `this.mask("(00) 0 0000-0000")`,
		},
		MASK_PHONE_BR: {
			JsCode: `this.mask("(00) 00000-0000");
let $this = this, 
	updateMask = function(event) {
		$this.off('blur');
		$this.unmask();
		if(this.value.replace(/\D/g, '').length > 10) {
			$this.mask("(00) 00000-0000");
		} else {
			$this.mask("(00) 0000-00009");
		}
		$(this).on('blur', updateMask);
	};
this.on('blur', updateMask);
`,
		},
	}
}
