# govalidator

A package of validators and sanitizers for strings, structs and collections. Based on [validator.js](https://github.com/chriso/validator.js).

This is a fork of https://github.com/asaskevich/govalidator to alter the behavior slightly:

- The new `Validate()` func returns all validation errors instead of just the first one found. The errors are returned in a map for easy JSON marshalling.
- The existing `ValidateStruct()` func remains unchanged from the original repo.
- Totally reworked the `README` to make it easier to understand the lib.

Things the original repo already did well include:

- Use of `valid` struct tags to define validations.
- Solid range of built in validators and the ability to create custom ones if required.
- Allows `optional` (tag) validation where validations are only run if the struct field value provided is non zero; essentially validating the value's correctness, not it's presence. For zero values, validation doesn't fail; because it's an optional field.
- Allows `required` (tag) validation where the struct field value provided must be non-zero based e.g. strings cannot be an empty string and integers cannot be zero etc.

## Installation

Install from the command line with:

    $ go get github.com/michaeltelford/govalidator

Import into your `*.go` files with:

```go
import "github.com/michaeltelford/govalidator"
```

## Usage

Below is an example `main.go` file validating a struct:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/michaeltelford/govalidator"
)

type User struct {
	Name  string `valid:"optional,length(2|20),in(Mick|Michael)" json:"name,omitempty"`
	Email string `valid:"email" json:"email,omitempty"`
	Age   int    `valid:"-" json:"age,omitempty"`
}

func main() {
	user := User{
		Name:  `M`,
		Email: `mic`,
		Age:   29,
	}

	valid, errs := govalidator.Validate(user)
	if !valid {
		json, _ := json.Marshal(errs)
		fmt.Println(string(json))
		return
	}

	fmt.Println(`valid`)
}
```

Run it via the command line with:

    $ go run main.go | jq

Which produces the following JSON errors (which could be returned as a 400 Bad Request in an API etc.):

```json
{
  "errors": {
    "email": [
      "mic does not validate as email"
    ],
    "name": [
      "M does not validate as length(2|20)",
      "M does not validate as in(Mick|Michael)"
    ]
  }
}
```

Things to note:

- Struct field `valid` tag contains list of comma separated validators.
- `govalidator.Validate(struct)` performs validations on the given struct.
- The returned `valid, errs` is of `bool, map` type for easy handling post validation.

### Built-in Validators

Here are the available built-in validators for use with struct fields:

```go
"email":              IsEmail,
"url":                IsURL,
"dialstring":         IsDialString,
"requrl":             IsRequestURL,
"requri":             IsRequestURI,
"alpha":              IsAlpha,
"utfletter":          IsUTFLetter,
"alphanum":           IsAlphanumeric,
"utfletternum":       IsUTFLetterNumeric,
"numeric":            IsNumeric,
"utfnumeric":         IsUTFNumeric,
"utfdigit":           IsUTFDigit,
"hexadecimal":        IsHexadecimal,
"hexcolor":           IsHexcolor,
"rgbcolor":           IsRGBcolor,
"lowercase":          IsLowerCase,
"uppercase":          IsUpperCase,
"int":                IsInt,
"float":              IsFloat,
"null":               IsNull,
"uuid":               IsUUID,
"uuidv3":             IsUUIDv3,
"uuidv4":             IsUUIDv4,
"uuidv5":             IsUUIDv5,
"creditcard":         IsCreditCard,
"isbn10":             IsISBN10,
"isbn13":             IsISBN13,
"json":               IsJSON,
"multibyte":          IsMultibyte,
"ascii":              IsASCII,
"printableascii":     IsPrintableASCII,
"fullwidth":          IsFullWidth,
"halfwidth":          IsHalfWidth,
"variablewidth":      IsVariableWidth,
"base64":             IsBase64,
"datauri":            IsDataURI,
"ip":                 IsIP,
"port":               IsPort,
"ipv4":               IsIPv4,
"ipv6":               IsIPv6,
"dns":                IsDNSName,
"host":               IsHost,
"mac":                IsMAC,
"latitude":           IsLatitude,
"longitude":          IsLongitude,
"ssn":                IsSSN,
"semver":             IsSemver,
"rfc3339":            IsRFC3339,
"rfc3339WithoutZone": IsRFC3339WithoutZone,
"ISO3166Alpha2":      IsISO3166Alpha2,
"ISO3166Alpha3":      IsISO3166Alpha3,
```

Built-in validators with parameters:

```go
"range(min|max)":                  Range,
"length(min|max)":                 ByteLength,
"runelength(min|max)":             RuneLength,
"matches(pattern)":                StringMatches,
"in(string1|string2|...|stringN)": IsIn,
```

## Advanced Usage

### Altering Default Validation Behavior

#### Validation Required

Activate behavior to require all fields have a validation tag by default.
`SetFieldsRequiredByDefault` causes validation to fail when struct fields do not include validations or are not explicitly marked as exempt (using `valid:"-"` or containing an `optional` tag e.g. `valid:"optional,email"`). This effectively applies the `valid:"required"` tag to all struct fields. A good place to activate this is a package `init()` function.

For example:

```go
import "github.com/michaeltelford/govalidator"

func init() {
  govalidator.SetFieldsRequiredByDefault(true)
}
```

Here's some code to help explain it:

```go
// this struct definition will fail govalidator.ValidateStruct() (and the field values do not matter):
type exampleStruct struct {
  Name  string ``
  Email string `valid:"email"`
}

// this, however, will only fail when Email is empty or an invalid email address:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"email"`
}

// lastly, this will only fail when Email is an invalid email address but not when it's empty:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"optional,email"`
}
```

#### Validation Error Messages

Custom error messages are supported via annotations by adding the `~` separator - here's an example of how to use it:

```go
type Ticket struct {
  Id        int64     `json:"id"`
  FirstName string    `json:"firstname" valid:"required~First name is blank"`
}
```

**Note**: Don't use colons (`:`) in your custom error messages as this affects the collection logic when working with all errors.

### Adding Custom Validators

Custom validation using your own domain specific validator tags is also available, here's an example of how to use it:

```go
package main

import "github.com/michaeltelford/govalidator"

type CustomByteArray [6]byte // custom types are supported and can be validated

type StructWithCustomByteArray struct {
  ID              CustomByteArray `valid:"customMinLengthValidator"` // custom tag.
  Email           string          `valid:"email"`
  CustomMinLength int             `valid:"-"`
}

func init() {
  govalidator.CustomTypeTagMap.Set("customMinLengthValidator", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
  switch v := context.(type) { // this validates a field against the value in another field, i.e. dependent validation
  case StructWithCustomByteArray:
    return len(v.ID) >= v.CustomMinLength
  }
  return false
  }))
}

func main() {
  s := StructWithCustomByteArray{
    ID:              CustomByteArray{1, 2, 3, 4, 5},
    Email:           `mick.telford@gmail.com`,
    CustomMinLength: 8,
  }

  valid, errs := govalidator.Validate(s)
  if !valid {
    json, _ := json.Marshal(errs)
    fmt.Println(string(json))
  } else {
    fmt.Println(`valid`)
  }
}
```

### Validation Functions

In addition to validation struct fields, you can validate single values as well using validation functions. It all works in the same way except there's no tag to link a field to a validator.

Below is an example:

```go
isValid := govalidator.IsURL(`http://user@pass:domain.com/path/page`)
```

#### Built-in Validator Functions

```go
func Abs(value float64) float64
func BlackList(str, chars string) string
func ByteLength(str string, params ...string) bool
func CamelCaseToUnderscore(str string) string
func Contains(str, substring string) bool
func Count(array []interface{}, iterator ConditionIterator) int
func Each(array []interface{}, iterator Iterator)
func ErrorByField(e error, field string) string
func ErrorsByField(e error) map[string]string
func Filter(array []interface{}, iterator ConditionIterator) []interface{}
func Find(array []interface{}, iterator ConditionIterator) interface{}
func GetLine(s string, index int) (string, error)
func GetLines(s string) []string
func InRange(value, left, right float64) bool
func IsASCII(str string) bool
func IsAlpha(str string) bool
func IsAlphanumeric(str string) bool
func IsBase64(str string) bool
func IsByteLength(str string, min, max int) bool
func IsCIDR(str string) bool
func IsCreditCard(str string) bool
func IsDNSName(str string) bool
func IsDataURI(str string) bool
func IsDialString(str string) bool
func IsDivisibleBy(str, num string) bool
func IsEmail(str string) bool
func IsFilePath(str string) (bool, int)
func IsFloat(str string) bool
func IsFullWidth(str string) bool
func IsHalfWidth(str string) bool
func IsHexadecimal(str string) bool
func IsHexcolor(str string) bool
func IsHost(str string) bool
func IsIP(str string) bool
func IsIPv4(str string) bool
func IsIPv6(str string) bool
func IsISBN(str string, version int) bool
func IsISBN10(str string) bool
func IsISBN13(str string) bool
func IsISO3166Alpha2(str string) bool
func IsISO3166Alpha3(str string) bool
func IsISO693Alpha2(str string) bool
func IsISO693Alpha3b(str string) bool
func IsISO4217(str string) bool
func IsIn(str string, params ...string) bool
func IsInt(str string) bool
func IsJSON(str string) bool
func IsLatitude(str string) bool
func IsLongitude(str string) bool
func IsLowerCase(str string) bool
func IsMAC(str string) bool
func IsMongoID(str string) bool
func IsMultibyte(str string) bool
func IsNatural(value float64) bool
func IsNegative(value float64) bool
func IsNonNegative(value float64) bool
func IsNonPositive(value float64) bool
func IsNull(str string) bool
func IsNumeric(str string) bool
func IsPort(str string) bool
func IsPositive(value float64) bool
func IsPrintableASCII(str string) bool
func IsRFC3339(str string) bool
func IsRFC3339WithoutZone(str string) bool
func IsRGBcolor(str string) bool
func IsRequestURI(rawurl string) bool
func IsRequestURL(rawurl string) bool
func IsSSN(str string) bool
func IsSemver(str string) bool
func IsTime(str string, format string) bool
func IsURL(str string) bool
func IsUTFDigit(str string) bool
func IsUTFLetter(str string) bool
func IsUTFLetterNumeric(str string) bool
func IsUTFNumeric(str string) bool
func IsUUID(str string) bool
func IsUUIDv3(str string) bool
func IsUUIDv4(str string) bool
func IsUUIDv5(str string) bool
func IsUpperCase(str string) bool
func IsVariableWidth(str string) bool
func IsWhole(value float64) bool
func LeftTrim(str, chars string) string
func Map(array []interface{}, iterator ResultIterator) []interface{}
func Matches(str, pattern string) bool
func NormalizeEmail(str string) (string, error)
func PadBoth(str string, padStr string, padLen int) string
func PadLeft(str string, padStr string, padLen int) string
func PadRight(str string, padStr string, padLen int) string
func Range(str string, params ...string) bool
func RemoveTags(s string) string
func ReplacePattern(str, pattern, replace string) string
func Reverse(s string) string
func RightTrim(str, chars string) string
func RuneLength(str string, params ...string) bool
func SafeFileName(str string) string
func SetFieldsRequiredByDefault(value bool)
func Sign(value float64) float64
func StringLength(str string, params ...string) bool
func StringMatches(s string, params ...string) bool
func StripLow(str string, keepNewLines bool) string
func ToBoolean(str string) (bool, error)
func ToFloat(str string) (float64, error)
func ToInt(str string) (int64, error)
func ToJSON(obj interface{}) (string, error)
func ToString(obj interface{}) string
func Trim(str, chars string) string
func Truncate(str string, length int, ending string) string
func UnderscoreToCamelCase(s string) string
func ValidateStruct(s interface{}) (bool, error)
func WhiteList(str, chars string) string
```
