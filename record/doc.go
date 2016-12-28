// Copyright (C) 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the terms of the Apache License 2.0

/*
Package record provides encoding and decoding utilities
for text files that store fixed-width records.

A fixed-width text file is a form of Flat file database [1],
commonly used to exchange data between
different financial institutions and some legacy systems,
usually with the intent to perform batch processing.


Encoding

The Encoder type converts struct fields
into fixed width records.
The struct fields of type string and int
are encoded with left-padding by default,
respectivelly using spaces or zeroes.
All other types are ignored by the encoder.

The convenience function Marshal
encodes a value and returns
the encoded record bytes.


Decoding

The Decoder type converts data from an io.Reader containing
fixed width values into a struct value.

You initialize the Decoder with a reader by calling NewDecoder,
and then call Decode() into a struct pointer
in order to convert a single line into a struct.

If all records within the reader are of the same type,
the same decoder can be used to read all data from the
reader, by issuing multiple calls to the Decode method.

All decoding errors are aggregated and returned as an
ErrorList, allowing callers to provide more usefull
error messages in their applications.


Tags

The struct fields can have a comma separated list of options
in a struct tag named `record`.

The tag can start with a number, like `record:"1"`,
that determines the size of the field.
String fields are padded with spaces to fill up to size,
and are truncated if higher than size.
Int fields are padded with zeroes to fill up to size.

The tag can have a "nopadding" option,
that avoids zero or space padding,
and uses the raw value as in %s and %d to fmt.Printf.
Note that this can make the resulting record variable in length.

The tag can have a "upper" option,
that causes strings to be upper cased before encoding.

The tag can have a "optional" value,
meaning that a decoding error due to empty value
will be silently ignored.
This tag does not affect invalid values like 'a'
for a decimal number.

When the tag option "-" is present, the field is skipped.


---

[1] http://en.wikipedia.org/wiki/Flat_file_database
*/
package record // import "ronoaldo.gopkg.net/encoding/record"
