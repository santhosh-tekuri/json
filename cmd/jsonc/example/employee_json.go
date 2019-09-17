// Code generated by jsonc; DO NOT EDIT.

package example

import "github.com/santhosh-tekuri/json"

func (e *employee) Unmarshal(de json.Decoder) error {
	return json.DecodeObj("employee", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Name"):
			if val := de.Token(); !val.Null() {
				e.Name, err = val.String("employee.Name")
			}
		case prop.Eq("sirName"):
			if val := de.Token(); !val.Null() {
				e.Sirname, err = val.String("employee.Sirname")
			}
		case prop.Eq("Permanent"):
			if val := de.Token(); !val.Null() {
				e.Permanent, err = val.Bool("employee.Permanent")
			}
		case prop.Eq("Height"):
			if val := de.Token(); !val.Null() {
				e.Height, err = val.Float64("employee.Height")
			}
		case prop.Eq("Weight"):
			if val := de.Token(); !val.Null() {
				e.Weight, err = val.Int("employee.Weight")
			}
		case prop.Eq("NickNames"):
			err = json.DecodeArr("employee.NickNames", de, func(de json.Decoder) error {
				item, err := de.Token().String("employee.NickNames[]")
				e.NickNames = append(e.NickNames, item)
				return err
			})
		case prop.Eq("Address"):
			err = e.Address.Unmarshal(de)
		case prop.Eq("Addresses"):
			err = json.DecodeArr("employee.Addresses", de, func(de json.Decoder) error {
				item := address{}
				err := item.Unmarshal(de)
				e.Addresses = append(e.Addresses, item)
				return err
			})
		case prop.Eq("Notes1"):
			e.Notes1, err = de.Decode()
		case prop.Eq("Notes2"):
			err = json.DecodeArr("employee.Notes2", de, func(de json.Decoder) error { item, err := de.Decode(); e.Notes2 = append(e.Notes2, item); return err })
		case prop.Eq("Notes3"):
			e.Notes3 = make(map[string]interface{})
			err = json.DecodeObj("employee.Notes3", de, func(de json.Decoder, prop json.Token) (err error) {
				k, _ := prop.String("")
				v, err := de.Decode()
				e.Notes3[k] = v
				return err
			})
		case prop.Eq("Contacts"):
			e.Contacts = make(map[string][]string)
			err = json.DecodeObj("employee.Contacts", de, func(de json.Decoder, prop json.Token) (err error) {
				k, _ := prop.String("")
				var v []string
				err = json.DecodeArr("employee.Contacts{}", de, func(de json.Decoder) error {
					item, err := de.Token().String("employee.Contacts{}[]")
					v = append(v, item)
					return err
				})
				e.Contacts[k] = v
				return err
			})
		case prop.Eq("Raw"):
			e.Raw, err = de.Marshal()
		case prop.Eq("Department"):
			err = json.DecodeObj("employee.Department", de, func(de json.Decoder, prop json.Token) (err error) {
				switch {
				case prop.Eq("name"):
					if val := de.Token(); !val.Null() {
						e.Department.Name, err = val.String("employee.Department.Name")
					}
				case prop.Eq("Manager"):
					if val := de.Token(); !val.Null() {
						e.Department.Manager, err = val.String("employee.Department.Manager")
					}
				default:
					err = de.Skip()
				}
				return
			})

		default:
			err = de.Skip()
		}
		return
	})
}
func (a *address) Unmarshal(de json.Decoder) error {
	return json.DecodeObj("address", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Street"):
			if val := de.Token(); !val.Null() {
				a.Street, err = val.String("address.Street")
			}
		default:
			err = de.Skip()
		}
		return
	})
}
