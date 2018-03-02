# File Name Template
This is a command line tool to replace {{keys}} inside of a file with their specified value. Can be used to replace a single template with seperate sets of values.

## Example
shellTemplate.sh
```bash
echo {{REPLACE_ME}}
echo {{AND_ME}}
```

### Single Replacement
`$ ./file_name_template shellTemplate.sh REPLACE_ME="Hello," AND_ME="World!"`

keyvalues.txt
```
REPLACE_ME="Hello,"
AND_ME="World!"
---
REPLACE_ME="Hey,"
AND_ME="Alice."
---
REPLACE_ME="Howdy,"
AND_ME="Bob."
```

### Multi Replacement
`$ cat keyvalues.txt | ./file_name_template shellTemplate.sh`
