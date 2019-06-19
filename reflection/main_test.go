package main

import "testing"

func TestAdditionalLocationArguments(t *testing.T) {
	tables := []struct {
		input            string
		args             additionalArguments
		strippedLocation string
		err              error
	}{
		{"/home/user", additionalArguments{}, "/home/user", nil},
		{"/home/user/", additionalArguments{}, "/home/user/", nil},
		{"/home/user/+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user", nil},
		{"/home/user+data", additionalArguments{}, "/home/user+data", nil},
		{"/home/user+data+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user+data", nil},
		{"/home/user+s/", additionalArguments{}, "/home/user+s/", nil},
		{"/home/user/+f", additionalArguments{firstLastPiecesFirst: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user/+sf", additionalArguments{sequentialDownload: ARGUMENT_TRUE, firstLastPiecesFirst: ARGUMENT_TRUE},
			"/home/user/", nil},
		{"/home/user/+h", additionalArguments{skipChecking: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user/-s", additionalArguments{sequentialDownload: ARGUMENT_FALSE}, "/home/user/", nil},
		{"/home/user/-sh", additionalArguments{sequentialDownload: ARGUMENT_FALSE, skipChecking:ARGUMENT_FALSE}, "/home/user/", nil},
		{"C:\\Users\\+s\\", additionalArguments{}, "C:\\Users\\+s\\", nil},
		{"C:\\Users\\+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "C:\\Users\\", nil},
	}

	for _, table := range tables {
		args, location, err := parseAdditionalLocationArguments(table.input)
		if args != table.args || location != table.strippedLocation || err != table.err {
			t.Errorf("Input %s, expected (%+v, %s, %v), got: (%+v, %s, %v)", table.input,
				table.args, table.strippedLocation, table.err,
				args, location, err)
		}
	}
}
