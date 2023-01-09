package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/qor5/admin/presets"
	"github.com/qor5/web"
	"github.com/qor5/web/multipartestutils"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/testingutils"
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
			Text(field.Value(obj).(time.Time).Format("2006-01-02")).
			Attr(web.VFieldName(field.Name)...)
	})

	ft.FieldType(Media("")).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		if field.ContextValue("a") == nil {
			return h.Text("")
		}
		return h.Text(field.ContextValue("a").(string) + ", " + field.ContextValue("b").(string))
	})

	r := httptest.NewRequest("GET", "/hello", nil)

	ctx := &web.EventContext{R: r, Flash: vd}

	time1 := time.Unix(1567048169, 0)
	time1LocalFormat := time1.Local().Format("2006-01-02 15:04:05")
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

	var cases = []struct {
		name           string
		toComponentFun func() h.HTMLComponent
		expect         string
	}{
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
<v-text-field type='number' v-field-name='[plaidForm, "Int1"]' label='整数1' :value='"2"' :disabled='false'></v-text-field>

<v-text-field type='number' v-field-name='[plaidForm, "Float1"]' label='Float1' :value='"23.1"' :disabled='false'></v-text-field>

<v-text-field type='text' v-field-name='[plaidForm, "String1"]' label='String1' :value='"hello"' :error-messages='["too small"]' :disabled='false'></v-text-field>

<v-checkbox v-field-name='[plaidForm, "Bool1"]' label='Bool1' :input-value='true' :disabled='false'></v-checkbox>

<div v-field-name='[plaidForm, "Time1"]' class='time-control'>2019-08-29</div>

<v-text-field type='text' v-field-name='[plaidForm, "Company.Name"]' label='公司名' :value='"Company1"' :disabled='false'></v-text-field>

<div v-field-name='[plaidForm, "Company.FoundedAt"]' class='time-control'>2019-08-29</div>
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
<v-text-field type='number' v-field-name='[plaidForm, "Int1"]' label='Int1' :value='"2"' :disabled='false'></v-text-field>

<v-text-field type='number' v-field-name='[plaidForm, "Float1"]' label='Float1' :value='"23.1"' :disabled='false'></v-text-field>

<v-text-field type='text' v-field-name='[plaidForm, "String1"]' label='String1' :value='"hello"' :error-messages='["too small"]' :disabled='false'></v-text-field>

<div v-field-name='[plaidForm, "Time1"]' class='time-control'>2019-08-29</div>
`,
		},

		{
			name: "Read Except with file glob pattern",
			toComponentFun: func() h.HTMLComponent {
				return ftRead.InspectFields(&User{}).
					Except("Float*").ToComponent(mb.Info(), user, ctx)
			},
			expect: `
<td>1</td>

<td>2</td>

<td>hello</td>

<td>true</td>
`,
		},

		{
			name: "Read for a time field",
			toComponentFun: func() h.HTMLComponent {
				return ftRead.InspectFields(&User{}).
					Only("Time1", "Int1").ToComponent(mb.Info(), user, ctx)
			},
			expect: fmt.Sprintf(`
<td>%s</td>

<td>2</td>
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
			expect: `context value1, context value2`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			output := h.MustString(c.toComponentFun(), web.WrapEventContext(context.TODO(), ctx))
			diff := testingutils.PrettyJsonDiff(c.expect, output)
			if len(diff) > 0 {
				t.Error(c.name, diff)
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
	return fmt.Sprintf(`<div>
<label class='v-label theme--light text-caption'>Address</label>

<v-card :outlined='true' class='mx-0 mt-1 mb-4 px-4 pb-0 pt-4'>
<v-text-field type='text' v-field-name='[plaidForm, "%sAddress.City"]' label='City' :value='"%s"' :disabled='false'></v-text-field>

<div>
<label class='v-label theme--light text-caption'>Detail</label>

<v-card :outlined='true' class='mx-0 mt-1 mb-4 px-4 pb-0 pt-4'>
<v-text-field type='text' v-field-name='[plaidForm, "%sAddress.Detail.Address1"]' label='Address1' :value='"%s"' :disabled='false'></v-text-field>

<v-text-field type='text' v-field-name='[plaidForm, "%sAddress.Detail.Address2"]' label='Address2' :value='"%s"' :disabled='false'></v-text-field>
</v-card>
</div>
</v-card>
</div>`,
		formKeyPrefix, v.City,
		formKeyPrefix, v.Detail.Address1,
		formKeyPrefix, v.Detail.Address2,
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

	var toComponentCases = []struct {
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
<input type='hidden' v-field-name='[plaidForm, "__Deleted.Departments[0].Employees"]' value='1,5'>

<input name='Name' type='text' value='Name 1'>

<div class='departments'>
<input name='Departments[0].Name' type='text' value='11111'>

<div class='employees'>
<input name='Departments[0].Employees[0].Number' type='text' value='111'>

%s

<input name='Departments[0].Employees[0].FakeNumber' type='text' value='900111'>

<input name='Departments[0].Employees[2].Number' type='text' value='333'>

%s

<input name='Departments[0].Employees[2].FakeNumber' type='text' value='900333'>

<button>Add Employee</button>
</div>

<input name='Departments[1].Name' type='text' value='22222'>

<div class='employees'>
<input name='Departments[1].Employees[0].Number' type='text' value='333'>

%s

<input name='Departments[1].Employees[0].FakeNumber' type='text' value='900333'>

<input name='Departments[1].Employees[1].Number' type='text' value='444'>

%s

<input name='Departments[1].Employees[1].FakeNumber' type='text' value='900444'>

<button>Add Employee</button>
</div>

<button>Add Department</button>
</div>

%s

<v-text-field type='number' v-field-name='[plaidForm, "PeopleCount"]' label='People Count' :value='"0"' :disabled='false'></v-text-field>
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
					SetSorted("Departments[0].Employees", []string{"2", "0", "3", "6"})
			},

			expectedHTML: fmt.Sprintf(`
<input type='hidden' v-field-name='[plaidForm, "__Deleted.Departments[0].Employees"]' value='1'>

<input type='hidden' v-field-name='[plaidForm, "__Sorted.Departments[0].Employees"]' value='2,0,3,6'>

<input name='Name' type='text' value='Name 1'>

<div class='departments'>
<input name='Departments[0].Name' type='text' value='11111'>

<div class='employees'>
<input name='Departments[0].Employees[2].Number' type='text' value='333'>

%s

<input name='Departments[0].Employees[2].FakeNumber' type='text' value='900333'>

<input name='Departments[0].Employees[0].Number' type='text' value='111'>

%s

<input name='Departments[0].Employees[0].FakeNumber' type='text' value='900111'>

<input name='Departments[0].Employees[3].Number' type='text' value='444'>

%s

<input name='Departments[0].Employees[3].FakeNumber' type='text' value='900444'>

<input name='Departments[0].Employees[4].Number' type='text' value='555'>

%s

<input name='Departments[0].Employees[4].FakeNumber' type='text' value='900555'>

<button>Add Employee</button>
</div>

<input name='Departments[1].Name' type='text' value='22222'>

<div class='employees'>
<input name='Departments[1].Employees[0].Number' type='text' value='333'>

%s

<input name='Departments[1].Employees[0].FakeNumber' type='text' value='900333'>

<input name='Departments[1].Employees[1].Number' type='text' value='444'>

%s

<input name='Departments[1].Employees[1].FakeNumber' type='text' value='900444'>

<button>Add Employee</button>
</div>

<button>Add Department</button>
</div>

%s

<v-text-field type='number' v-field-name='[plaidForm, "PeopleCount"]' label='People Count' :value='"0"' :disabled='false'></v-text-field>
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
				R: httptest.NewRequest("POST", "/", nil),
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

	var unmarshalCases = []struct {
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
				AddField("__Deleted.Departments[0].Employees", "1,2").
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
						Name: "Department C",
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
