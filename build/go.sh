# ------------------------------------ #
# Tool to set up Go when using Cygwin  #
# ------------------------------------ #
# Make sure to change the Go installation 
# path below if necessary:
GOINSTALLPATH=C:/Go
# Usage: @nasa-space-rocks-bot $ source build/go.sh
export GOROOT=$(cygpath -m ${GOINSTALLPATH%?}); echo GOROOT=$GOROOT
export GOPATH=$(cygpath -m ${PWD}); echo GOPATH=$GOPATH
# export GOARCH=amd64; echo GOARCH=$GOARCH
export PATH=${GOINSTALLPATH%?}/bin:${PATH}