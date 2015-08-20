# Command dsfinfo

* Remove the encode from the end of main()
* Add examples

# Package audio

* Re-write to use an interface e.g. audio.Audio should probably be an interface rather than a struct

# Package dsf

* Create writeDataChunk() and hook into encode()
* Migrate the remaining existing tests from https://github.com/snmoore/go-dsd.git and rework into the new style
* Investigate getting the table driven tests to generate individual TestXxx functions
    * Failures are hard to follow because each test is really just an iteration of a loop within TestDsdChunk() etc
    * Would be better if each test is an individual test function
    * So want to generate and return test functions from the test tables
* Add examples

# Package dff

* Migrate from https://github.com/snmoore/go-dsd.git and rework into the new style as per package dsf

# Miscellaneous

* Reconsider the use of the decoder pattern borrowed from image.Image
    * It does not quite feel right...