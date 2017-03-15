# tangotandas

Go and python implementation of TangoTandas plugin for iTunes

### Description
For tango DJ's it is valuable to show the audience what current tanda is and
what is coming next. This simple plugin allows to do that. It scan current
playlist from iTunes, figure out which tanda is playing and show it either
on standard output or dump output in HTML format.

The convention are the following:
- the Tanda is a set of Tangos, Vals or Milongas tracks
- the supported genres are Tango, Vals, Milonga and Cortina
- the current track will be display in color (red for Tango, green for Vals
purple for Milonga and cyan for Cortina). Also either a note or cup of
coffee unicode symbol will be displayed in front of it.

