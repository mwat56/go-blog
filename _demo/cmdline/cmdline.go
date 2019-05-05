/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
               All rights reserved
           EMail : <support@mwat.de>
*/
package main

import (
	"fmt"

	"github.com/mwat56/go-blog"
)

func main() {
	fmt.Printf("\n%d entries:\n%v\n\n", len(blog.AppArguments), blog.AppArguments)
} // main()

/* _EoF_ */
