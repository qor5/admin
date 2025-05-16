package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/testingutils"

	. "github.com/qor5/admin/v3/presets"
)

type Company struct {
	Name      string
	FoundedAt time.Time
}

type Media string

type User struct {
	ID      int
	Int1    int
	Float1  float32
	String1 string
	Bool1   bool
	Time1   time.Time
	Company *Company
	Media1  Media
}

func TestFields(t *testing.T) {
	vd := &web.ValidationErrors{}
	vd.FieldError("String1", "too small")

	ft := NewFieldDefaults(WRITE).Exclude("ID")
	ft.FieldType(time.Time{}).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div().Class("time-control").
			Attr(web.VField(field.Name, field.Value(obj).(time.Time).Format("2006-01-02"))...)
	})

	ft.FieldType(Media("")).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		if field.ContextValue("a") == nil {
			return h.Text("")
		}
		return h.Text(field.ContextValue("a").(string) + ", " + field.ContextValue("b").(string))
	})

	r := httptest.NewRequest("GET", "/hello", http.NoBody)

	ctx := &web.EventContext{R: r, Flash: vd}

	time1 := time.Unix(1567048169, 0)
	time1LocalFormat := time1.Local().Format("2006-01-02 15:04:05")
	time1LocalFormatMinute := time1.Local().Format("2006-01-02 15:04")
	user := &User{
		ID:      1,
		Int1:    2,
		Float1:  23.1,
		String1: "hello",
		Bool1:   true,
		Time1:   time1,
		Company: &Company{
			Name:      "Company1",
			FoundedAt: time.Unix(1567048169, 0),
		},
	}
	mb := New().Model(&User{})

	ftRead := NewFieldDefaults(LIST)

	type testCase struct {
		name           string
		toComponentFun func() h.HTMLComponent
		expect         string
	}

	cases := []testCase{
		{
			name: "creating should copy editing",
			toComponentFun: func() h.HTMLComponent {
				ed := mb.Editing("Int1", "Float1")
				ed.Field("Float1").ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
					return h.Div().Class("my_float32").Text(fmt.Sprintf("%f", field.Value(obj).(float32)))
				})
				creating := ed.Creating().Except("Int1")
				return creating.FieldsBuilder.ToComponent(mb.Info(), user, ctx)
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Float1"]===undefined || dash.visible["Float1"]'>
<div class='my_float32'>23.100000</div>
</div>
`,
		},
		{
			name: "Only with additional nested object",
			toComponentFun: func() h.HTMLComponent {
				return ft.InspectFields(&User{}).
					Labels("Int1", "整数1", "Company.Name", "公司名").
					Only("Int1", "Float1", "String1", "Bool1", "Time1", "Company.Name", "Company.FoundedAt").
					ToComponent(
						mb.Info(),
						user,
						ctx)
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Int1"]===undefined || dash.visible["Int1"]'>
<vx-field type='number' v-model='form["Int1"]' :error-messages='dash.errorMessages["Int1"]' v-assign:append='[dash.errorMessages, {"Int1":null}]' v-assign='[form, {"Int1":"2"}]' label='整数1' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["Float1"]===undefined || dash.visible["Float1"]'>
<vx-field type='number' v-model='form["Float1"]' :error-messages='dash.errorMessages["Float1"]' v-assign:append='[dash.errorMessages, {"Float1":null}]' v-assign='[form, {"Float1":"23.1"}]' label='Float1' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["String1"]===undefined || dash.visible["String1"]'>
<vx-field label='String1' v-model='form["String1"]' :error-messages='dash.errorMessages["String1"]' v-assign:append='[dash.errorMessages, {"String1":["too small"]}]' v-assign='[form, {"String1":"hello"}]' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["Bool1"]===undefined || dash.visible["Bool1"]'>
<vx-checkbox v-model='form["Bool1"]' :error-messages='dash.errorMessages["Bool1"]' v-assign:append='[dash.errorMessages, {"Bool1":null}]' v-assign='[form, {"Bool1":true}]' label='Bool1' :disabled='false'></vx-checkbox>
</div>

<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<div v-model='form["Time1"]' v-assign='[form, {"Time1":"2019-08-29"}]' class='time-control'></div>
</div>

<div v-show='!dash.visible || dash.visible["Company.Name"]===undefined || dash.visible["Company.Name"]'>
<vx-field label='公司名' v-model='form["Company.Name"]' :error-messages='dash.errorMessages["Company.Name"]' v-assign:append='[dash.errorMessages, {"Company.Name":null}]' v-assign='[form, {"Company.Name":"Company1"}]' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["Company.FoundedAt"]===undefined || dash.visible["Company.FoundedAt"]'>
<div v-model='form["Company.FoundedAt"]' v-assign='[form, {"Company.FoundedAt":"2019-08-29"}]' class='time-control'></div>
</div>
`,
		},

		{
			name: "Except with file glob pattern",
			toComponentFun: func() h.HTMLComponent {
				return ft.InspectFields(&User{}).
					Except("Bool*").
					ToComponent(mb.Info(), user, ctx)
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Int1"]===undefined || dash.visible["Int1"]'>
<vx-field type='number' v-model='form["Int1"]' :error-messages='dash.errorMessages["Int1"]' v-assign:append='[dash.errorMessages, {"Int1":null}]' v-assign='[form, {"Int1":"2"}]' label='Int1' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["Float1"]===undefined || dash.visible["Float1"]'>
<vx-field type='number' v-model='form["Float1"]' :error-messages='dash.errorMessages["Float1"]' v-assign:append='[dash.errorMessages, {"Float1":null}]' v-assign='[form, {"Float1":"23.1"}]' label='Float1' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["String1"]===undefined || dash.visible["String1"]'>
<vx-field label='String1' v-model='form["String1"]' :error-messages='dash.errorMessages["String1"]' v-assign:append='[dash.errorMessages, {"String1":["too small"]}]' v-assign='[form, {"String1":"hello"}]' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<div v-model='form["Time1"]' v-assign='[form, {"Time1":"2019-08-29"}]' class='time-control'></div>
</div>

<div v-show='!dash.visible || dash.visible["Media1"]===undefined || dash.visible["Media1"]'></div>
`,
		},

		{
			name: "Read Except with file glob pattern",
			toComponentFun: func() h.HTMLComponent {
				return ftRead.InspectFields(&User{}).
					Except("Float*").ToComponent(mb.Info(), user, ctx)
			},
			expect: fmt.Sprintf(`
<div v-show='!dash.visible || dash.visible["ID"]===undefined || dash.visible["ID"]'>
<td>1</td>
</div>

<div v-show='!dash.visible || dash.visible["Int1"]===undefined || dash.visible["Int1"]'>
<td>2</td>
</div>

<div v-show='!dash.visible || dash.visible["String1"]===undefined || dash.visible["String1"]'>
<td>hello</td>
</div>

<div v-show='!dash.visible || dash.visible["Bool1"]===undefined || dash.visible["Bool1"]'>
<td>true</td>
</div>

<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<td>%s</td>
</div>
`, time1LocalFormat),
		},

		{
			name: "Read for a time field",
			toComponentFun: func() h.HTMLComponent {
				return ftRead.InspectFields(&User{}).
					Only("Time1", "Int1").ToComponent(mb.Info(), user, ctx)
			},
			expect: fmt.Sprintf(`
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<td>%s</td>
</div>

<div v-show='!dash.visible || dash.visible["Int1"]===undefined || dash.visible["Int1"]'>
<td>2</td>
</div>
`, time1LocalFormat),
		},

		{
			name: "pass in context",
			toComponentFun: func() h.HTMLComponent {
				fb := ft.InspectFields(&User{}).
					Only("Media1")
				fb.Field("Media1").
					WithContextValue("a", "context value1").
					WithContextValue("b", "context value2")
				return fb.ToComponent(mb.Info(), user, ctx)
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Media1"]===undefined || dash.visible["Media1"]'>context value1, context value2</div>
`,
		},
		{
			name: "Display default time component",
			toComponentFun: func() h.HTMLComponent {
				vd := &web.ValidationErrors{}
				vd.FieldError("Time1", "err time1")

				return NewFieldDefaults(WRITE).Exclude("ID").InspectFields(&User{}).
					Labels("Company.FoundedAt", "公司创立于").
					Only("Time1", "Company.FoundedAt").
					ToComponent(
						mb.Info(),
						user,
						&web.EventContext{R: r, Flash: vd})
			},
			expect: fmt.Sprintf(`
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='Time1' v-model='form["Time1"]' :error-messages='dash.errorMessages["Time1"]' v-assign:append='[dash.errorMessages, {"Time1":["err time1"]}]' v-assign='[form, {"Time1":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>

<div v-show='!dash.visible || dash.visible["Company.FoundedAt"]===undefined || dash.visible["Company.FoundedAt"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='公司创立于' v-model='form["Company.FoundedAt"]' :error-messages='dash.errorMessages["Company.FoundedAt"]' v-assign:append='[dash.errorMessages, {"Company.FoundedAt":null}]' v-assign='[form, {"Company.FoundedAt":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>
`, time1LocalFormatMinute, time1LocalFormatMinute),
		},
		{
			name: "Display zero time with default time component",
			toComponentFun: func() h.HTMLComponent {
				vd := &web.ValidationErrors{}
				vd.FieldError("Time1", "err time1")

				type WithTime struct {
					Time1 time.Time
					Time2 *time.Time
				}
				return NewFieldDefaults(WRITE).InspectFields(&WithTime{}).ToComponent(
					mb.Info(),
					&WithTime{
						Time1: time.Time{},
						Time2: &time.Time{},
					},
					&web.EventContext{R: r, Flash: vd})
			},
			expect: fmt.Sprintf(`
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='Time1' v-model='form["Time1"]' :error-messages='dash.errorMessages["Time1"]' v-assign:append='[dash.errorMessages, {"Time1":["err time1"]}]' v-assign='[form, {"Time1":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>

<div v-show='!dash.visible || dash.visible["Time2"]===undefined || dash.visible["Time2"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='Time2' v-model='form["Time2"]' :error-messages='dash.errorMessages["Time2"]' v-assign:append='[dash.errorMessages, {"Time2":null}]' v-assign='[form, {"Time2":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>
`, "", ""),
		},
		{
			name: "Display nil time with default time component",
			toComponentFun: func() h.HTMLComponent {
				vd := &web.ValidationErrors{}
				vd.FieldError("Time1", "err time1")

				type WithTime struct {
					Time1 time.Time
					Time2 *time.Time
				}
				return NewFieldDefaults(WRITE).InspectFields(&WithTime{}).ToComponent(
					mb.Info(),
					&WithTime{
						Time1: time.Time{},
						Time2: nil,
					},
					&web.EventContext{R: r, Flash: vd})
			},
			expect: fmt.Sprintf(`
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='Time1' v-model='form["Time1"]' :error-messages='dash.errorMessages["Time1"]' v-assign:append='[dash.errorMessages, {"Time1":["err time1"]}]' v-assign='[form, {"Time1":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>

<div v-show='!dash.visible || dash.visible["Time2"]===undefined || dash.visible["Time2"]'>
<vx-date-picker type='datetimepicker' format='YYYY-MM-DD HH:mm' label='Time2' v-model='form["Time2"]' :error-messages='dash.errorMessages["Time2"]' v-assign:append='[dash.errorMessages, {"Time2":null}]' v-assign='[form, {"Time2":%q}]' :clearable='true' :disabled='false'></vx-date-picker>
</div>
`, "", ""),
		},
		{
			name: "Display zero time with default time component(READ)",
			toComponentFun: func() h.HTMLComponent {
				vd := &web.ValidationErrors{}
				vd.FieldError("Time1", "err time1")

				type WithTime struct {
					Time1 time.Time
					Time2 *time.Time
				}
				return NewFieldDefaults(DETAIL).InspectFields(&WithTime{}).ToComponent(
					mb.Info(),
					&WithTime{
						Time1: time.Time{},
						Time2: &time.Time{},
					},
					&web.EventContext{R: r, Flash: vd})
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<div class='mb-4'>
<label class='v-label theme--light text-caption'>Time1</label>

<div class='pt-1'></div>
</div>
</div>

<div v-show='!dash.visible || dash.visible["Time2"]===undefined || dash.visible["Time2"]'>
<div class='mb-4'>
<label class='v-label theme--light text-caption'>Time2</label>

<div class='pt-1'></div>
</div>
</div>
`,
		},
		{
			name: "Display nil time with default time component(READ)",
			toComponentFun: func() h.HTMLComponent {
				vd := &web.ValidationErrors{}
				vd.FieldError("Time1", "err time1")

				type WithTime struct {
					Time1 time.Time
					Time2 *time.Time
				}
				return NewFieldDefaults(DETAIL).InspectFields(&WithTime{}).ToComponent(
					mb.Info(),
					&WithTime{
						Time1: time.Time{},
						Time2: nil,
					},
					&web.EventContext{R: r, Flash: vd})
			},
			expect: `
<div v-show='!dash.visible || dash.visible["Time1"]===undefined || dash.visible["Time1"]'>
<div class='mb-4'>
<label class='v-label theme--light text-caption'>Time1</label>

<div class='pt-1'></div>
</div>
</div>

<div v-show='!dash.visible || dash.visible["Time2"]===undefined || dash.visible["Time2"]'>
<div class='mb-4'>
<label class='v-label theme--light text-caption'>Time2</label>

<div class='pt-1'>&lt;nil&gt;</div>
</div>
</div>
`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			output := h.MustString(c.toComponentFun(), web.WrapEventContext(context.TODO(), ctx))
			diff := testingutils.PrettyJsonDiff(c.expect, output)
			if diff != "" {
				t.Error(c.name, diff)
				t.Logf("\nexpected: %s\noutput: %s", c.expect, output)
			}
		})
	}
}

type Person struct {
	Addresses []*Org
}

type Org struct {
	Name        string
	Address     Address
	PeopleCount int
	Departments []*Department
}

type Department struct {
	Name      string
	Employees []*Employee
	DBStatus  string
}

type Employee struct {
	Number  int
	Address *Address
}

type Address struct {
	City   string
	Detail AddressDetail
}

type AddressDetail struct {
	Address1 string
	Address2 string
}

func addressHTML(v Address, formKeyPrefix string) string {
	return fmt.Sprintf(`<div v-show='!dash.visible || dash.visible["%sAddress"]===undefined || dash.visible["%sAddress"]'>
<div>
<label class='v-label theme--light text-caption wrapper-field-label'>Address</label>

<v-card :variant='"outlined"' class='mx-0 mt-1 mb-4 px-4 pb-0 pt-4'>
<div v-show='!dash.visible || dash.visible["%sAddress.City"]===undefined || dash.visible["%sAddress.City"]'>
<vx-field label='City' v-model='form["%sAddress.City"]' :error-messages='dash.errorMessages["%sAddress.City"]' v-assign:append='[dash.errorMessages, {"%sAddress.City":null}]' v-assign='[form, {"%sAddress.City":"%s"}]' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["%sAddress.Detail"]===undefined || dash.visible["%sAddress.Detail"]'>
<div>
<label class='v-label theme--light text-caption wrapper-field-label'>Detail</label>

<v-card :variant='"outlined"' class='mx-0 mt-1 mb-4 px-4 pb-0 pt-4'>
<div v-show='!dash.visible || dash.visible["%sAddress.Detail.Address1"]===undefined || dash.visible["%sAddress.Detail.Address1"]'>
<vx-field label='Address1' v-model='form["%sAddress.Detail.Address1"]' :error-messages='dash.errorMessages["%sAddress.Detail.Address1"]' v-assign:append='[dash.errorMessages, {"%sAddress.Detail.Address1":null}]' v-assign='[form, {"%sAddress.Detail.Address1":"%s"}]' :disabled='false'></vx-field>
</div>

<div v-show='!dash.visible || dash.visible["%sAddress.Detail.Address2"]===undefined || dash.visible["%sAddress.Detail.Address2"]'>
<vx-field label='Address2' v-model='form["%sAddress.Detail.Address2"]' :error-messages='dash.errorMessages["%sAddress.Detail.Address2"]' v-assign:append='[dash.errorMessages, {"%sAddress.Detail.Address2":null}]' v-assign='[form, {"%sAddress.Detail.Address2":"%s"}]' :disabled='false'></vx-field>
</div>
</v-card>
</div>
</div>
</v-card>
</div>
</div>`,
		formKeyPrefix, formKeyPrefix,
		formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, v.City,
		formKeyPrefix, formKeyPrefix,
		formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, v.Detail.Address1,
		formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, formKeyPrefix, v.Detail.Address2,
	)
}

func TestFieldsBuilder(t *testing.T) {
	defaults := NewFieldDefaults(WRITE)

	addressFb := NewFieldsBuilder().Model(&Address{}).Defaults(defaults).Only("City", "Detail")
	addressDetailFb := NewFieldsBuilder().Model(&AddressDetail{}).Defaults(defaults).Only("Address1", "Address2")
	addressFb.Field("Detail").Nested(addressDetailFb)

	employeeFbs := NewFieldsBuilder().Model(&Employee{}).Defaults(defaults)
	employeeFbs.Field("Number").ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Input(field.FormKey).Type("text").Value(field.StringValue(obj))
	})

	employeeFbs.Field("Address").Nested(addressFb)

	employeeFbs.Field("FakeNumber").ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Input(field.FormKey).Type("text").Value(fmt.Sprintf("900%v", reflectutils.MustGet(obj, "Number")))
	}).SetterFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.FormValue(field.FormKey)
		if v == "" {
			return
		}
		return reflectutils.Set(obj, "Number", "900"+v)
	})

	deptFbs := NewFieldsBuilder().Model(&Department{}).Defaults(defaults)
	deptFbs.Field("Name").ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// [0].Departments[0].Name
		// [0].Departments[1].Name
		// [1].Departments[0].Name
		return h.Input(field.FormKey).Type("text").Value(field.StringValue(obj))
	}).SetterFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
		reflectutils.Set(obj, field.Name, ctx.R.FormValue(field.FormKey)+"!!!")
		// panic("dept name setter")
		return
	})

	deptFbs.Field("Employees").Nested(employeeFbs).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(
			field.NestedFieldsBuilder.ToComponentForEach(field, obj.(*Department).Employees, ctx, nil),
			h.Button("Add Employee"),
		).Class("employees")
	})

	fbs := NewFieldsBuilder().Model(&Org{}).Defaults(defaults)
	fbs.Field("Name").ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// [0].Name
		return h.Input(field.Name).Type("text").Value(field.StringValue(obj))
	})
	// .SetterFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
	// 	return
	// })

	fbs.Field("Departments").Nested(deptFbs).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// [0].Departments
		return h.Div(
			field.NestedFieldsBuilder.ToComponentForEach(field, obj.(*Org).Departments, ctx, nil),
			h.Button("Add Department"),
		).Class("departments")
	})

	fbs.Field("Address").Nested(addressFb)

	fbs.Field("PeopleCount").SetterFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
		reflectutils.Set(obj, field.Name, ctx.R.FormValue(field.FormKey))
		return
	})

	toComponentCases := []struct {
		name         string
		obj          *Org
		setup        func(ctx *web.EventContext)
		expectedHTML string
	}{
		{
			name: "Only deleted",
			obj: &Org{
				Name: "Name 1",
				Address: Address{
					City: "c1",
					Detail: AddressDetail{
						Address1: "addr1",
						Address2: "addr2",
					},
				},
				Departments: []*Department{
					{
						Name: "11111",
						Employees: []*Employee{
							{Number: 111},
							{Number: 222},
							{Number: 333},
						},
					},
					{
						Name: "22222",
						Employees: []*Employee{
							{Number: 333},
							{Number: 444},
						},
					},
				},
			},
			setup: func(ctx *web.EventContext) {
				ContextModifiedIndexesBuilder(ctx).
					AppendDeleted("Departments[0].Employees", 1).
					AppendDeleted("Departments[0].Employees", 5)
			},

			expectedHTML: fmt.Sprintf(`
<input type='hidden' v-model='form["__Deleted.Departments[0].Employees"]' v-assign='[form, {"__Deleted.Departments[0].Employees":"1,5"}]'>

<div v-show='!dash.visible || dash.visible["Name"]===undefined || dash.visible["Name"]'>
<input name='Name' type='text' value='Name 1'>
</div>

<div v-show='!dash.visible || dash.visible["Departments"]===undefined || dash.visible["Departments"]'>
<div class='departments'>
<div v-show='!dash.visible || dash.visible["Departments[0].Name"]===undefined || dash.visible["Departments[0].Name"]'>
<input name='Departments[0].Name' type='text' value='11111'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees"]===undefined || dash.visible["Departments[0].Employees"]'>
<div class='employees'>
<div v-show='!dash.visible || dash.visible["Departments[0].Employees[0].Number"]===undefined || dash.visible["Departments[0].Employees[0].Number"]'>
<input name='Departments[0].Employees[0].Number' type='text' value='111'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[0].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[0].FakeNumber"]'>
<input name='Departments[0].Employees[0].FakeNumber' type='text' value='900111'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[2].Number"]===undefined || dash.visible["Departments[0].Employees[2].Number"]'>
<input name='Departments[0].Employees[2].Number' type='text' value='333'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[2].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[2].FakeNumber"]'>
<input name='Departments[0].Employees[2].FakeNumber' type='text' value='900333'>
</div>

<button>Add Employee</button>
</div>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Name"]===undefined || dash.visible["Departments[1].Name"]'>
<input name='Departments[1].Name' type='text' value='22222'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Employees"]===undefined || dash.visible["Departments[1].Employees"]'>
<div class='employees'>
<div v-show='!dash.visible || dash.visible["Departments[1].Employees[0].Number"]===undefined || dash.visible["Departments[1].Employees[0].Number"]'>
<input name='Departments[1].Employees[0].Number' type='text' value='333'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[0].FakeNumber"]===undefined || dash.visible["Departments[1].Employees[0].FakeNumber"]'>
<input name='Departments[1].Employees[0].FakeNumber' type='text' value='900333'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[1].Number"]===undefined || dash.visible["Departments[1].Employees[1].Number"]'>
<input name='Departments[1].Employees[1].Number' type='text' value='444'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[1].FakeNumber"]===undefined || dash.visible["Departments[1].Employees[1].FakeNumber"]'>
<input name='Departments[1].Employees[1].FakeNumber' type='text' value='900444'>
</div>

<button>Add Employee</button>
</div>
</div>

<button>Add Department</button>
</div>
</div>

%s

<div v-show='!dash.visible || dash.visible["PeopleCount"]===undefined || dash.visible["PeopleCount"]'>
<vx-field type='number' v-model='form["PeopleCount"]' :error-messages='dash.errorMessages["PeopleCount"]' v-assign:append='[dash.errorMessages, {"PeopleCount":null}]' v-assign='[form, {"PeopleCount":"0"}]' label='People Count' :disabled='false'></vx-field>
</div>
`,
				addressHTML(Address{}, "Departments[0].Employees[0]."),
				addressHTML(Address{}, "Departments[0].Employees[2]."),
				addressHTML(Address{}, "Departments[1].Employees[0]."),
				addressHTML(Address{}, "Departments[1].Employees[1]."),
				addressHTML(Address{
					City: "c1",
					Detail: AddressDetail{
						Address1: "addr1",
						Address2: "addr2",
					},
				}, "")),
		},

		{
			name: "Deleted with Sorted",
			obj: &Org{
				Name: "Name 1",
				Departments: []*Department{
					{
						Name: "11111",
						Employees: []*Employee{
							{Number: 111},
							{Number: 222},
							{Number: 333},
							{Number: 444},
							{Number: 555},
						},
					},
					{
						Name: "22222",
						Employees: []*Employee{
							{Number: 333},
							{Number: 444},
						},
					},
				},
			},
			setup: func(ctx *web.EventContext) {
				ContextModifiedIndexesBuilder(ctx).
					AppendDeleted("Departments[0].Employees", 1).
					Sorted("Departments[0].Employees", []string{"2", "0", "3", "6"})
			},

			expectedHTML: fmt.Sprintf(`
<input type='hidden' v-model='form["__Deleted.Departments[0].Employees"]' v-assign='[form, {"__Deleted.Departments[0].Employees":"1"}]'>

<input type='hidden' v-model='form["__Sorted.Departments[0].Employees"]' v-assign='[form, {"__Sorted.Departments[0].Employees":"2,0,3,6"}]'>

<div v-show='!dash.visible || dash.visible["Name"]===undefined || dash.visible["Name"]'>
<input name='Name' type='text' value='Name 1'>
</div>

<div v-show='!dash.visible || dash.visible["Departments"]===undefined || dash.visible["Departments"]'>
<div class='departments'>
<div v-show='!dash.visible || dash.visible["Departments[0].Name"]===undefined || dash.visible["Departments[0].Name"]'>
<input name='Departments[0].Name' type='text' value='11111'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees"]===undefined || dash.visible["Departments[0].Employees"]'>
<div class='employees'>
<div v-show='!dash.visible || dash.visible["Departments[0].Employees[2].Number"]===undefined || dash.visible["Departments[0].Employees[2].Number"]'>
<input name='Departments[0].Employees[2].Number' type='text' value='333'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[2].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[2].FakeNumber"]'>
<input name='Departments[0].Employees[2].FakeNumber' type='text' value='900333'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[0].Number"]===undefined || dash.visible["Departments[0].Employees[0].Number"]'>
<input name='Departments[0].Employees[0].Number' type='text' value='111'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[0].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[0].FakeNumber"]'>
<input name='Departments[0].Employees[0].FakeNumber' type='text' value='900111'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[3].Number"]===undefined || dash.visible["Departments[0].Employees[3].Number"]'>
<input name='Departments[0].Employees[3].Number' type='text' value='444'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[3].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[3].FakeNumber"]'>
<input name='Departments[0].Employees[3].FakeNumber' type='text' value='900444'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[4].Number"]===undefined || dash.visible["Departments[0].Employees[4].Number"]'>
<input name='Departments[0].Employees[4].Number' type='text' value='555'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[0].Employees[4].FakeNumber"]===undefined || dash.visible["Departments[0].Employees[4].FakeNumber"]'>
<input name='Departments[0].Employees[4].FakeNumber' type='text' value='900555'>
</div>

<button>Add Employee</button>
</div>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Name"]===undefined || dash.visible["Departments[1].Name"]'>
<input name='Departments[1].Name' type='text' value='22222'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Employees"]===undefined || dash.visible["Departments[1].Employees"]'>
<div class='employees'>
<div v-show='!dash.visible || dash.visible["Departments[1].Employees[0].Number"]===undefined || dash.visible["Departments[1].Employees[0].Number"]'>
<input name='Departments[1].Employees[0].Number' type='text' value='333'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[0].FakeNumber"]===undefined || dash.visible["Departments[1].Employees[0].FakeNumber"]'>
<input name='Departments[1].Employees[0].FakeNumber' type='text' value='900333'>
</div>

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[1].Number"]===undefined || dash.visible["Departments[1].Employees[1].Number"]'>
<input name='Departments[1].Employees[1].Number' type='text' value='444'>
</div>

%s

<div v-show='!dash.visible || dash.visible["Departments[1].Employees[1].FakeNumber"]===undefined || dash.visible["Departments[1].Employees[1].FakeNumber"]'>
<input name='Departments[1].Employees[1].FakeNumber' type='text' value='900444'>
</div>

<button>Add Employee</button>
</div>
</div>

<button>Add Department</button>
</div>
</div>

%s

<div v-show='!dash.visible || dash.visible["PeopleCount"]===undefined || dash.visible["PeopleCount"]'>
<vx-field type='number' v-model='form["PeopleCount"]' :error-messages='dash.errorMessages["PeopleCount"]' v-assign:append='[dash.errorMessages, {"PeopleCount":null}]' v-assign='[form, {"PeopleCount":"0"}]' label='People Count' :disabled='false'></vx-field>
</div>
`,
				addressHTML(Address{}, "Departments[0].Employees[2]."),
				addressHTML(Address{}, "Departments[0].Employees[0]."),
				addressHTML(Address{}, "Departments[0].Employees[3]."),
				addressHTML(Address{}, "Departments[0].Employees[4]."),
				addressHTML(Address{}, "Departments[1].Employees[0]."),
				addressHTML(Address{}, "Departments[1].Employees[1]."),
				addressHTML(Address{}, ""),
			),
		},
	}

	for _, c := range toComponentCases {
		t.Run(c.name, func(t *testing.T) {
			ctx := &web.EventContext{
				R: httptest.NewRequest("POST", "/", http.NoBody),
			}
			c.setup(ctx)
			result := fbs.ToComponent(nil, c.obj, ctx)
			actual1 := h.MustString(result, context.TODO())

			diff := testingutils.PrettyJsonDiff(c.expectedHTML, actual1)
			if diff != "" {
				t.Error(diff)
			}
		})
	}

	unmarshalCases := []struct {
		name                 string
		initial              *Org
		expected             *Org
		req                  *http.Request
		removeDeletedAndSort bool
	}{
		{
			name: "case with deleted",
			initial: &Org{
				Departments: []*Department{
					{
						Name: "Department A",
						Employees: []*Employee{
							{
								Number: 0,
							},
							{
								Number: 1,
							},
						},
					},
					{
						Name: "Department B",
						Employees: []*Employee{
							{Number: 0},
							{Number: 1},
							{Number: 2},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
			req: multipartestutils.NewMultipartBuilder().
				AddField("Name", "Org 1").
				AddField("PeopleCount", "420").
				AddField("Departments[1].Name", "Department 1").
				AddField("Departments[1].Employees[0].Number", "888").
				AddField("Departments[1].Employees[2].Number", "999").
				AddField("Departments[1].Employees[1].FakeNumber", "666").
				AddField("Departments[1].DBStatus", "Verified").
				AddField("Departments[2].Name", "Department C").
				AddField("__Deleted.Departments[0].Employees", "1,2").
				AddField("__Deleted.Departments[1].Employees", "0").
				AddField("__Deleted.Departments[2].Employees", "0").
				BuildEventFuncRequest(),
			removeDeletedAndSort: false,
			expected: &Org{
				Name:        "Org 1",
				PeopleCount: 420,
				Departments: []*Department{
					{
						Name: "!!!",
						Employees: []*Employee{
							{
								Number: 0,
							},
							nil,
							nil,
						},
					},
					{
						Name: "Department 1!!!",
						Employees: []*Employee{
							nil,
							{
								Number:  900666,
								Address: &Address{},
							},
							{
								Number:  999,
								Address: &Address{},
							},
						},
					},
					{
						Name: "Department C!!!",
						Employees: []*Employee{
							nil,
						},
					},
				},
			},
		},

		{
			name: "removeDeletedAndSort true",
			initial: &Org{
				Departments: []*Department{
					{
						Name: "Department A",
						Employees: []*Employee{
							{
								Number: 0,
							},
							{
								Number: 1,
							},
						},
					},
					{
						Name: "Department B",
						Employees: []*Employee{
							{Number: 0},
							{Number: 1},
							{Number: 2},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
			req: multipartestutils.NewMultipartBuilder().
				AddField("Name", "Org 1").
				AddField("PeopleCount", "420").
				AddField("Departments[1].Name", "Department 1").
				AddField("Departments[1].Employees[0].Number", "888").
				AddField("Departments[1].Employees[2].Number", "999").
				AddField("Departments[1].Employees[1].FakeNumber", "666").
				AddField("Departments[1].DBStatus", "Verified").
				AddField("__Deleted.Departments[0].Employees", "1,5").
				AddField("__Deleted.Departments[1].Employees", "0").
				BuildEventFuncRequest(),
			removeDeletedAndSort: true,
			expected: &Org{
				Name:        "Org 1",
				PeopleCount: 420,
				Departments: []*Department{
					{
						Name: "!!!",
						Employees: []*Employee{
							{
								Number: 0,
							},
						},
					},
					{
						Name: "Department 1!!!",
						Employees: []*Employee{
							{
								Number:  900666,
								Address: &Address{},
							},
							{
								Number:  999,
								Address: &Address{},
							},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
		},
		{
			name: "removeDeletedAndSort true and sorted",
			initial: &Org{
				Departments: []*Department{
					{
						Name: "Department A",
						Employees: []*Employee{
							{
								Number: 0,
							},
							{
								Number: 1,
							},
							{
								Number: 2,
							},
						},
					},
					{
						Name: "Department B",
						Employees: []*Employee{
							{Number: 0},
							{Number: 1},
							{Number: 2},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
			req: multipartestutils.NewMultipartBuilder().
				AddField("Name", "Org 1").
				AddField("PeopleCount", "420").
				AddField("Departments[1].Name", "Department 1").
				AddField("Departments[1].Employees[0].Number", "000").
				AddField("Departments[1].Employees[2].Number", "222").
				AddField("Departments[1].Employees[3].Number", "333").
				AddField("Departments[1].Employees[4].Number", "444").
				AddField("Departments[1].Employees[1].FakeNumber", "666"). // this will set Number[1] to 900666
				AddField("Departments[1].DBStatus", "Verified").
				AddField("__Deleted.Departments[0].Employees", "1,5").
				AddField("__Sorted.Departments[1].Employees", "3,4,0,2").
				AddField("__Deleted.Departments[1].Employees", "0").
				BuildEventFuncRequest(),
			removeDeletedAndSort: true,
			expected: &Org{
				Name:        "Org 1",
				PeopleCount: 420,
				Departments: []*Department{
					{
						Name: "!!!",
						Employees: []*Employee{
							{
								Number: 0,
							},
							{
								Number: 2,
							},
						},
					},
					{
						Name: "Department 1!!!",
						Employees: []*Employee{
							{
								Number:  333,
								Address: &Address{},
							},
							{
								Number:  444,
								Address: &Address{},
							},
							{
								Number:  222,
								Address: &Address{},
							},
							{
								Number:  900666,
								Address: &Address{},
							},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
		},

		{
			name: "removeDeletedAndSort false and sorted",
			initial: &Org{
				Departments: []*Department{
					{
						Name: "Department A",
						Employees: []*Employee{
							{
								Number: 0,
							},
							{
								Number: 1,
							},
							{
								Number: 2,
							},
						},
					},
					{
						Name: "Department B",
						Employees: []*Employee{
							{Number: 0},
							{Number: 1},
							{Number: 2},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
			req: multipartestutils.NewMultipartBuilder().
				AddField("Name", "Org 1").
				AddField("PeopleCount", "420").
				AddField("Departments[1].Name", "Department 1").
				AddField("Departments[1].Employees[0].Number", "000").
				AddField("Departments[1].Employees[2].Number", "222").
				AddField("Departments[1].Employees[3].Number", "333").
				AddField("Departments[1].Employees[4].Number", "444").
				AddField("Departments[1].Employees[1].FakeNumber", "666"). // this will set Number[1] to 900666
				AddField("Departments[1].DBStatus", "Verified").
				AddField("__Deleted.Departments[0].Employees", "1,3").
				AddField("__Sorted.Departments[1].Employees", "3,4,0,2").
				AddField("__Deleted.Departments[1].Employees", "0").
				BuildEventFuncRequest(),
			removeDeletedAndSort: false,
			expected: &Org{
				Name:        "Org 1",
				PeopleCount: 420,
				Departments: []*Department{
					{
						Name: "!!!",
						Employees: []*Employee{
							{
								Number: 0,
							},
							nil,
							{
								Number: 2,
							},
							nil,
						},
					},
					{
						Name: "Department 1!!!",
						Employees: []*Employee{
							nil,
							{
								Number:  900666,
								Address: &Address{},
							},
							{
								Number:  222,
								Address: &Address{},
							},
							{
								Number:  333,
								Address: &Address{},
							},
							{
								Number:  444,
								Address: &Address{},
							},
						},
					},
					{
						Name: "Department C",
					},
				},
			},
		},

		{
			name: "object item",
			initial: &Org{
				Name: "Org 1",
				Address: Address{
					City: "org city",
					Detail: AddressDetail{
						Address1: "org addr1",
						Address2: "org addr2",
					},
				},
				Departments: []*Department{
					{
						Name: "Department A",
						Employees: []*Employee{
							{
								Number: 1,
								Address: &Address{
									City: "1 city",
									Detail: AddressDetail{
										Address1: "1 addr1",
										Address2: "1 addr2",
									},
								},
							},
							{
								Number: 2,
							},
						},
					},
				},
			},
			req: multipartestutils.NewMultipartBuilder().
				AddField("Name", "Org 1").
				AddField("Address.City", "org city e").
				AddField("Address.Detail.Address1", "org addr1 e").
				AddField("Address.Detail.Address2", "org addr2 e").
				AddField("Departments[0].Name", "Department A").
				AddField("Departments[0].Employees[0].Number", "1").
				AddField("Departments[0].Employees[0].Address.City", "1 city e").
				AddField("Departments[0].Employees[0].Address.Detail.Address1", "1 addr1 e").
				AddField("Departments[0].Employees[0].Address.Detail.Address2", "1 addr2 e").
				AddField("Departments[0].Employees[1].Number", "2").
				AddField("Departments[0].Employees[1].Address.City", "2 city").
				AddField("Departments[0].Employees[1].Address.Detail.Address1", "2 addr1").
				AddField("Departments[0].Employees[1].Address.Detail.Address2", "2 addr2").
				BuildEventFuncRequest(),
			removeDeletedAndSort: false,
			expected: &Org{
				Name: "Org 1",
				Address: Address{
					City: "org city e",
					Detail: AddressDetail{
						Address1: "org addr1 e",
						Address2: "org addr2 e",
					},
				},
				Departments: []*Department{
					{
						Name: "Department A!!!",
						Employees: []*Employee{
							{
								Number: 1,
								Address: &Address{
									City: "1 city e",
									Detail: AddressDetail{
										Address1: "1 addr1 e",
										Address2: "1 addr2 e",
									},
								},
							},
							{
								Number: 2,
								Address: &Address{
									City: "2 city",
									Detail: AddressDetail{
										Address1: "2 addr1",
										Address2: "2 addr2",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range unmarshalCases {
		t.Run(c.name, func(t *testing.T) {
			ctx2 := &web.EventContext{R: c.req}
			_ = ctx2.R.ParseMultipartForm(128 << 20)
			actual2 := c.initial
			vErr := fbs.Unmarshal(actual2, nil, c.removeDeletedAndSort, ctx2)
			if vErr.HaveErrors() {
				t.Error(vErr.Error())
			}
			diff := testingutils.PrettyJsonDiff(c.expected, actual2)
			if diff != "" {
				t.Error(diff)
			}
		})
	}
}
