package field_test

import (
	"context"
	"os"
	"testing"

	"github.com/qor5/admin/v3/presets/field"
	"github.com/qor5/web/v3"
)

type Org struct {
	Name        string
	Address     Address
	PeopleCount int
	Departments []*Department
	MultiNested [][][]Address
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

var tplOrg = &Org{
	Name: "Example Org",
	Address: Address{
		City: "Example City",
		Detail: AddressDetail{
			Address1: "123 Example St",
			Address2: "Suite 456",
		},
	},
	PeopleCount: 100,
	Departments: []*Department{
		{
			Name: "HR",
			Employees: []*Employee{
				{Number: 1, Address: &Address{City: "City A"}},
				{Number: 2, Address: &Address{City: "City B"}},
			},
			DBStatus: "Active",
		},
		{
			Name: "Engineering",
			Employees: []*Employee{
				{Number: 3, Address: &Address{City: "City C"}},
				{Number: 4, Address: &Address{City: "City D"}},
			},
			DBStatus: "Active",
		},
	},
	MultiNested: [][][]Address{
		{
			{
				{
					City: "(MultiNested0) City",
					Detail: AddressDetail{
						Address1: "(MultiNested0) 123 Example St",
						Address2: "(MultiNested0) Suite 456",
					},
				},
				{
					City: "(MultiNested1) City",
					Detail: AddressDetail{
						Address1: "(MultiNested1) 123 Example St",
						Address2: "(MultiNested1) Suite 456",
					},
				},
			},
		},
	},
}

func TestStructBuilder(t *testing.T) {
	// 给予样本用以构造
	fb := field.Scope(field.Inspect(&Org{}), "/orgs/qor5")
	// 给到具体的 obj
	compo, err := fb.Build(&web.EventContext{}, &field.Context{Value: tplOrg})
	if err != nil {
		panic(err)
	}
	data, err := compo.MarshalHTML(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 先只写入到文件里便于格式化查看
	if err := os.WriteFile("./_x.html", data, 0o644); err != nil {
		t.Fatal(err)
	}
}
