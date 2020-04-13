# GoSlicer

This is a very experimental slicer for 3d printing.

The initial work of GoSlicer is based on the first CuraEngine commits.
As I had no clue where to start, I chose to port the initial Cura commit to Go.
The code of this early Cura version already provides a very simple and working slicer and the code of it is easy to read.
https://github.com/Ultimaker/CuraEngine/tree/80dc349e2014eaa9450086c007118e10bda0b534

## Run
go run /path/to/stl/file.stl

##
* ~~read stl~~ (initially done by using external lib github.com/hschendel/stl)
* ~~implement optimisation as in first Cura Commit~~
* first slicing result
* simple infill
* bottom / top layer
* options as parameters / config file (using cobra / viper)
* lots of other things...