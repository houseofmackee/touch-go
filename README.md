# touch-go
Simple implementation of the touch command in Go.

The code is a bit meh, but this is my first "proper" tool written in Go, so I'm ok with it.

Just so you know, while the code has the logic to get and use the last access time from a reference file, it doesn't have the actual code to do so. It will take the last modified time from the reference file and use it for both the modified time stamp and the last accessed time stamp.

Usage: touch [options] files...

The supported command line options are:

**-a**    change only the access time (not properly implemented yet)

**-m**    change only the modification time

**-c**, **--no-create**
      do not create any files

**-R**, **--recursive**
      process all files and directories inside directories passed as arguments

**-r**, **--reference=FILE**
      use this file's times instead of the current time

**-h**, **--help**
      display a short help text and exit

**-v**, **--version**
      display version info and exit
