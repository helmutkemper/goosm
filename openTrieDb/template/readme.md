# Binary template

This file is a C++ template for [010 Editor](https://www.sweetscape.com/010editor)

```C++
struct FILE {
    struct LINE {
		uint length<bgcolor=cLtGreen>;
		uint64 nextAddress<bgcolor=cLtBlue>;
		struct SLOT {
			byte piece[4]<bgcolor=cLtRed>;
			int64 addressLine<bgcolor=cLtYellow>;
			byte flag<bgcolor=cLtGray, comment=FlagName>;
		} dataSet[64]<format=hex, fgcolor=cBlack, bgcolor=cLtGray>;    
	} lineData[9999];
} file;

string FlagName(byte &flag) {
    if( flag==0x00 ){ return "flagNotSet"; }
    if( flag==0x01 ){ return "flagContinue"; }
    if( flag==0x02 ){ return "flagComplete"; }
    if( flag==0x03 ){ return "flagContComp"; }
    if( flag==0x04 ){ return "flagItSelf"; }
   
    return "";
}
```

On line `lineData[9999]`, the template expects `9999` lines of data. Just adjust the value.
