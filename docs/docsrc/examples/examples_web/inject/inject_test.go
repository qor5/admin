package inject

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	injector := New()
	require.NotNil(t, injector)
	require.NotNil(t, injector.values)
	require.NotNil(t, injector.providers)
}

func TestSetParent(t *testing.T) {
	injector := New()
	parent := New()
	err := injector.SetParent(parent)
	require.NoError(t, err)
	require.Equal(t, parent, injector.parent)
	err = injector.SetParent(parent)
	require.ErrorIs(t, err, ErrParentAlreadySet)
}

func TestProvide(t *testing.T) {
	{
		injector := New()
		require.Panics(t, func() {
			injector.Provide("testNotFunc")
		})
	}
	{
		injector := New()
		err := injector.Provide(func() string { return "test" })
		require.NoError(t, err)
	}
	{
		injector := New()
		err := injector.Provide(func() (string, int) { return "test", 0 })
		require.NoError(t, err)
		require.Len(t, injector.providers, 3)

		err = injector.Provide(func() (int64, int) { return 1, 2 })
		require.ErrorIs(t, err, ErrTypeAlreadyProvided)
		require.Len(t, injector.providers, 3)
	}
	{
		injector := New()
		err := injector.Provide(func() (string, error) { return "test", nil })
		require.NoError(t, err)
		require.Len(t, injector.providers, 2)

		_, err = injector.invoke(func(s string) {})
		require.NoError(t, err)
		require.Len(t, injector.providers, 1)

		err = injector.Provide(func() (int64, string) { return 1, "" })
		require.ErrorIs(t, err, ErrTypeAlreadyProvided)
		require.Len(t, injector.providers, 1)
	}
}

func TestMustProvide(t *testing.T) {
	injector := New()
	require.NotPanics(t, func() {
		injector.MustProvide(func() (string, int) { return "test", 0 })
	})
	require.Panics(t, func() {
		injector.MustProvide(func() string { return "test" })
	})
}

func TestInvoke(t *testing.T) {
	injector := New()
	err := injector.Provide(func() string { return "test" })
	require.NoError(t, err)

	require.Panics(t, func() {
		injector.invoke("testNotFunc")
	})

	results, err := injector.invoke(func(s string) string { return s })
	require.NoError(t, err)
	require.Equal(t, "test", results[0].Interface())

	{
		errTemp := errors.New("temp")
		results, err := injector.invoke(func(s string) (string, error) { return "", errTemp })
		require.ErrorIs(t, err, errTemp)
		require.Equal(t, "", results[0].Interface())
		require.Len(t, results, 1)
	}

	{
		results, err := injector.Invoke(func(s string) string { return s })
		require.NoError(t, err)
		require.Equal(t, "test", results[0])
	}
}

func TestResolve(t *testing.T) {
	injector := New()
	err := injector.Provide(func() string { return "test" })
	require.NoError(t, err)
	value, err := injector.resolve(reflect.TypeOf(""))
	require.NoError(t, err)
	require.Equal(t, "test", value.Interface())

	{
		var str string
		err = injector.Resolve(&str)
		require.NoError(t, err)
		require.Equal(t, "test", str)
	}

	err = injector.Provide(func() *string {
		a := "testPtr"
		return &a
	})
	require.NoError(t, err)

	{
		var str *string
		err = injector.Resolve(&str)
		require.NoError(t, err)
		require.Equal(t, "testPtr", *str)
	}

	{
		injector := New()

		errTemp := errors.New("temp")
		// if error is not nil, the value will be ignored
		err := injector.Provide(func() (string, error) { return "test", errTemp })
		require.NoError(t, err)
		{
			value, err := injector.resolve(reflect.TypeOf(""))
			require.ErrorIs(t, err, errTemp)
			require.True(t, !value.IsValid())
		}
		{
			str := "xxx"
			err = injector.Resolve(&str)
			require.ErrorIs(t, err, errTemp)
			require.Equal(t, "xxx", str)
		}
	}
}

func TestApply(t *testing.T) {
	type TestStruct struct {
		*Injector `inject:""`
		Value     string `inject:"" json:"value,omitempty"`
		value     string `inject:""`
		optional0 *int64 `inject:"optional"`
		optional1 uint64 `inject:"optional"`
		ID        string `json:"id,omitempty"`
	}
	injector := New()
	err := injector.Provide(
		func() string { return "test" },
		func() uint64 { return 123 },
	)
	require.NoError(t, err)
	testStruct := &TestStruct{}
	err = injector.Apply(testStruct)
	require.NoError(t, err)
	require.Equal(t, "test", testStruct.Value)
	require.Equal(t, "test", testStruct.value)
	require.Nil(t, testStruct.optional0)
	require.Equal(t, uint64(123), testStruct.optional1)
	require.Equal(t, "", testStruct.ID)
	require.Equal(t, injector, testStruct.Injector)

	require.Panics(t, func() {
		injector.Apply("testNotStruct")
	})
}

func TestMultipleProviders(t *testing.T) {
	injector := New()
	err := injector.Provide(func() string { return "test1" })
	require.NoError(t, err)
	err = injector.Provide(func() string { return "test2" })
	require.ErrorIs(t, err, ErrTypeAlreadyProvided)
	results, err := injector.invoke(func(s1, s2 string) string { return s1 + s2 })
	require.NoError(t, err)
	require.Equal(t, "test1test1", results[0].Interface())
}

func TestUnresolvedDependency(t *testing.T) {
	injector := New()
	err := injector.Provide(func() string { return "test" })
	require.NoError(t, err)
	_, err = injector.invoke(func(s string, i int) string { return s })
	require.ErrorIs(t, err, ErrTypeNotProvided)
}

func TestParentInjection(t *testing.T) {
	parent := New()
	err := parent.Provide(func() string { return "test" })
	require.NoError(t, err)
	child := New()
	err = child.SetParent(parent)
	require.NoError(t, err)
	results, err := child.invoke(func(s string) string { return s })
	require.NoError(t, err)
	require.Equal(t, "test", results[0].Interface())

	// override
	err = child.Provide(func() string { return "test2" })
	require.NoError(t, err)
	results, err = child.invoke(func(s string) string { return s })
	require.NoError(t, err)
	require.Equal(t, "test2", results[0].Interface())
}

type TestInterface interface {
	Test() string
}

type TestStruct struct {
	Name string
}

func (t *TestStruct) Test() string {
	return t.Name
}

func TestInterfaceType(t *testing.T) {
	injector := New()
	err := injector.Provide(func() TestInterface { return &TestStruct{Name: "hello"} })
	require.NoError(t, err)
	value, err := injector.resolve(reflect.TypeOf((*TestInterface)(nil)).Elem())
	require.NoError(t, err)
	require.NotNil(t, value.Interface())
	require.Equal(t, "hello", value.Interface().(TestInterface).Test())

	type Visibility string
	err = injector.Provide(func() Visibility { return "public" })
	require.NoError(t, err)
	value, err = injector.resolve(reflect.TypeOf(Visibility("")))
	require.NoError(t, err)
	require.Equal(t, Visibility("public"), value.Interface())

	type StructToApply struct {
		iface      TestInterface `inject:""`
		Visibility Visibility    `inject:""`
		str        string        `inject:""`
		ID         string        `json:"id,omitempty"`
	}
	err = injector.Provide(func() string { return "str" })
	require.NoError(t, err)

	structToApply := &StructToApply{}
	err = injector.Apply(structToApply)
	require.NoError(t, err)
	require.Equal(t, "hello", structToApply.iface.Test())
	require.Equal(t, Visibility("public"), structToApply.Visibility)
	require.Equal(t, "str", structToApply.str)
	require.Equal(t, "", structToApply.ID)
}

func TestAutoApply(t *testing.T) {
	type TestStruct struct {
		Injector *Injector `inject:""`
		Value    string    `inject:"" json:"value,omitempty"`
		value    string    `inject:""`
		ID       string    `json:"id,omitempty"`
	}
	injector := New()
	err := injector.Provide(
		func() string { return "test" },
	)
	require.NoError(t, err)
	results, err := injector.Invoke(func() *TestStruct {
		return &TestStruct{
			ID: "testID",
		}
	})
	require.NoError(t, err)
	testStruct := results[0].(*TestStruct)
	require.Equal(t, "test", testStruct.Value)
	require.Equal(t, "test", testStruct.value)
	require.Equal(t, "testID", testStruct.ID)
	require.Equal(t, injector, testStruct.Injector)

	{
		err = injector.Provide(func() *TestStruct { return &TestStruct{ID: "testID2"} })
		require.NoError(t, err)

		testStruct := &TestStruct{}
		err := injector.Resolve(&testStruct)
		require.NoError(t, err)
		require.Equal(t, "test", testStruct.Value)
		require.Equal(t, "test", testStruct.value)
		require.Equal(t, "testID2", testStruct.ID)
		require.Equal(t, injector, testStruct.Injector)
	}
}
