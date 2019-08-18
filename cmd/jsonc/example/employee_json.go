// Code generated by jsonc; DO NOT EDIT.

package example

import "github.com/santhosh-tekuri/json"

func (e *employee) Unmarshal(de json.Decoder) error {
	return json.UnmarshalObj("employee", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Name"):
			e.Name, err = de.Token().String("employee.Name")
		case prop.Eq("sirName"):
			e.Sirname, err = de.Token().String("employee.Sirname")
		case prop.Eq("Permanent"):
			e.Permanent, err = de.Token().Bool("employee.Permanent")
		case prop.Eq("Height"):
			e.Height, err = de.Token().Float64("employee.Height")
		case prop.Eq("Weight"):
			e.Weight, err = de.Token().Int("employee.Weight")
		case prop.Eq("NickNames"):
			err = json.UnmarshalArr("employee.NickNames", de, func(de json.Decoder) error {
				item, err := de.Token().String("employee.NickNames[]")
				e.NickNames = append(e.NickNames, item)
				return err
			})
		case prop.Eq("Address"):
			err = e.Address.Unmarshal(de)
		case prop.Eq("Addresses"):
			err = json.UnmarshalArr("employee.Addresses", de, func(de json.Decoder) error {
				item := address{}
				err := item.Unmarshal(de)
				e.Addresses = append(e.Addresses, item)
				return err
			})
		case prop.Eq("Notes1"):
			e.Notes1, err = de.Unmarshal()
		case prop.Eq("Notes2"):
			err = json.UnmarshalArr("employee.Notes2", de, func(de json.Decoder) error {
				item, err := de.Unmarshal()
				e.Notes2 = append(e.Notes2, item)
				return err
			})
		default:
			err = de.Skip()
		}
		return
	})
}

func (a *address) Unmarshal(de json.Decoder) error {
	return json.UnmarshalObj("address", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("Street"):
			a.Street, err = de.Token().String("address.Street")
		default:
			err = de.Skip()
		}
		return
	})
}
