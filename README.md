# govalidator

A package of validators and sanitizers for strings, structs and collections. Based on [validator.js](https://github.com/chriso/validator.js). Perfect for validating HTTP API's.

This is a fork of https://github.com/asaskevich/govalidator to alter and extend the behavior. Things the original repo already did well include:

- Use of `valid:""` struct tags to define field validations.
- Solid range of built in validators and the ability to create custom ones if required.
- Allows `required` (tag) validation where the struct field value provided must be non-zero based e.g. strings cannot be an empty string and integers cannot be zero etc.

New changes since forking the repo:

- The new `Validate()` func returns all validation errors instead of just the first one found. The errors are returned in a `map` for easy processing e.g. JSON marshalling. When arrays/slices are found, they are traversed until the first invalid element (of type struct) is found and all of its field's errors are returned.
- Added new `valid` tags such as `forbidden` and `nonemptystring` etc.
- Totally reworked the `README` to make it easier to use the lib.

## Installation

Install from the command line with:

    $ go get gopkg.in/michaeltelford/govalidator.v11

Import into your `*.go` files with:

```go
import "gopkg.in/michaeltelford/govalidator.v11"
```

Then refer to the package as `govalidator`.

## Docs

View the `godocs` at:

https://godoc.org/gopkg.in/michaeltelford/govalidator.v11

## Basic Usage

Below is an example `main.go` file validating a struct:

```go
package main

import (
	"encoding/json"
	"fmt"

	"gopkg.in/michaeltelford/govalidator.v11"
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

Which produces the following JSON errors (which could be returned as a 400 Bad Request in an API response etc.):

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
- `govalidator.Validate(user)` performs validations on the given struct.
- Only public fields are validated, private fields are skipped.
- The returned `valid, errs` is of `bool, map` types for easy handling post validation.
- The returned `errs` contain the `json` tag field names (if provided).

### Built-in Validators

#### Validating Field Presence

If you are validating when a struct field value is present (or absent) you can use the following validator tags inside `valid:""`:

| Tag             | Description |
| --------------- | ----------- |
| `-`             | No validations are performed. |
| `optional`      | To be used with other validators (separated by a comma e.g. `optional,email`). Run all other validators if value is non zero, otherwise skip this field. |
| `forbidden` | A field must have a zero value set. |
| `required`      | A field must have a non zero value set. Note that `required` isn't needed with other validators that inheritantly validate a value's presence e.g. `nonemptystring`. Omitting `required` in these cases reduces the number of error messages. |

#### Validating String Values

If you are validating a `string` struct field's value then below are some of the most common validation tags:

| Tag             | Description |
| --------------- | ----------- |
| `nonemptystring` | A field must have a non empty string value set. Whitespace is trimmed. |
| `numeric` | A field must have a string value containing an integer. If valid, a string to int conversion will succeed. |
| `boolean` | A field must have a string value containing a boolean. If valid, a string to boolean conversion will succeed. |

#### Validating Field Correctness

Below is the full list of built-in validator tags for validating the  correctness of a struct field value; multiple field types are supported:

```go
"nonemptystring":     IsNonEmptyString
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
"boolean":            IsBoolean,
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

#### Custom Error Messages

Custom error messages are supported via annotations by adding the `~` separator - here's an example of how to use it:

```go
type Ticket struct {
  Id        int64     `json:"id" valid:"-"`
  FirstName string    `json:"firstname" valid:"required~First name is blank"`
}
```

#### Requiring Rules By Default

Activate behavior to require all fields have a validation tag by default.
`SetFieldsRequiredByDefault` causes validation to fail when struct fields do not include validations or are not explicitly marked as exempt (using `valid:"-"` or containing an `optional` tag e.g. `valid:"optional,email"`). This effectively applies the `valid:"required"` tag to all struct fields. A good place to activate this is a package `init()` function.

For example:

```go
import "gopkg.in/michaeltelford/govalidator.v11"

func init() {
  govalidator.SetFieldsRequiredByDefault(true)
}
```

Here's some code to help explain it:

```go
// This struct definition will fail govalidator.Validate() (and the field values do not matter):
type exampleStruct struct {
  Name  string ``
  Email string `valid:"email"`
}

// This, however, will only fail when Email is empty or an invalid email address:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"email"`
}

// Lastly, this will only fail when Email is an invalid email address but not when it's empty:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"optional,email"`
}
```

### Adding Custom Validators

Custom validation using your own domain specific validator tags is also available, here's a (somewhat advanced) example of how to use it:

```go
package main

import "gopkg.in/michaeltelford/govalidator.v11"

type CustomByteArray [6]byte // Custom types are supported and can be validated.

type StructWithCustomByteArray struct {
  ID              CustomByteArray `valid:"customMinLengthValidator"` // Custom tag.
  Email           string          `valid:"email"`
  CustomMinLength int             `valid:"-"`
}

func init() {
  govalidator.CustomTypeTagMap.Set("customMinLengthValidator", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
  switch v := context.(type) { // This validates a field against the value in another field, i.e. dependent validation.
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

In addition to validating struct fields, you can validate single values as well using validation functions. It all works in the same way except there's no tag linking a field to a validator.

Below is an example:

```go
isValid := govalidator.IsURL(`http://user@pass:domain.com/path/page`)
```

Here is another which validates that a URL's ID field is a numeric value:

```go
// Pass in user ID string and attribute name (used in errs).
// id will be 0 if conversion fails.
id, errs := govalidator.ConvertToInt(userIDStr, `user_id`)
if id < 1 {
  // Invalid user ID, use errs map...
}
// Conversion succeeded, use id (of type int) as needed...
```

#### Built-in Validator Functions

```go
func ConvertToInt(str, attr string) (int, map[string]map[string][]string)
func ConvertToBool(str, attr string) (bool, map[string]map[string][]string, error)
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
func IsBoolean(str string) bool
func IsByteLength(str string, min, max int) bool
func IsCIDR(str string) bool
func IsCreditCard(str string) bool
func IsDNSName(str string) bool
func IsDataURI(str string) bool
func IsDialString(str string) bool
func IsDivisibleBy(str, num string) bool
func IsEmail(str string) bool
func IsEmptyString(str string) bool
func IsNonEmptyString(str string) bool
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
func WhiteList(str, chars string) string
```
