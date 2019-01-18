# YAML Parser
YAML parser parses a YAML file to a map[string]interface{} and then traverses the map and adds values into the config store flattening the YAML using dot notation for keys. 

Ex: 
```
foo: "bar"
nested: 
    firstName: "john"
    lastName: "doe"
    list: 
    - 1
    - 2
```
Will add the following key/value to the config
```
"foo" => "bar"
"nested.firstName" => "john"
"nested.lastName" => "doe"
"nested.list" => []int{1,2}
```

# Usage
```
err := kpyaml.Parser.Parse(strings.NewReader(`foo: "bar"`), konfig.Values{})
```
