// Code generated by jsonc; DO NOT EDIT.

package tests

import "github.com/santhosh-tekuri/json"

func (s *stringVal) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("stringVal", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			if val := de.Token(); !val.Null() {
				s.Field, err = val.String("stringVal.Field")
			}
		default:
			err = de.Skip()
		}
		return
	})
}
func (s *structTag) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("structTag", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Name"):
			if val := de.Token(); !val.Null() {
				s.Field, err = val.String("structTag.Field")
			}
		default:
			err = de.Skip()
		}
		return
	})
}
func (e *excludeTag) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("excludeTag", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {

		default:
			err = de.Skip()
		}
		return
	})
}
func (u *unexported) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("unexported", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {

		default:
			err = de.Skip()
		}
		return
	})
}
func (a *arrString) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("arrString", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			err = json.DecodeArr("arrString.Field", de, func(de json.Decoder) error {
				item, err := de.Token().String("arrString.Field[]")
				a.Field = append(a.Field, item)
				return err
			})
		default:
			err = de.Skip()
		}
		return
	})
}
func (p *ptrString) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("ptrString", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			p.Field = nil
			if val := de.Token(); !val.Null() {
				var pval string
				pval, err = val.String("ptrString.Field")
				p.Field = &pval
			}
		default:
			err = de.Skip()
		}
		return
	})
}
func (a *arrPtrString) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("arrPtrString", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			err = json.DecodeArr("arrPtrString.Field", de, func(de json.Decoder) error {
				var item *string
				var err error
				if val := de.Token(); !val.Null() {
					var pval string
					pval, err = val.String("arrPtrString.Field[]")
					item = &pval
				}
				a.Field = append(a.Field, item)
				return err
			})
		default:
			err = de.Skip()
		}
		return
	})
}
func (i *interfaceVal) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("interfaceVal", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			i.Field, err = de.Decode()
		default:
			err = de.Skip()
		}
		return
	})
}
func (a *arrInterface) DecodeJSON(de json.Decoder) error {
	return json.DecodeObj("arrInterface", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Field"):
			err = json.DecodeArr("arrInterface.Field", de, func(de json.Decoder) error {
				item, err := de.Decode()
				a.Field = append(a.Field, item)
				return err
			})
		default:
			err = de.Skip()
		}
		return
	})
}
