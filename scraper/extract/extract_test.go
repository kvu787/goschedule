package extract

import (
	"fmt"
	"testing"
)

const (
	sampleClassDescriptionIndex string = `<P><B><A NAME="astr101">ASTR 101 </A> Astronomy (5) NW, QSR </B><BR>Introduction to the universe, with emphasis on conceptual, as contrasted with mathematical, comprehension. Modern theories, observations; ideas concerning nature, evolution of galaxies; quasars, stars, black holes, planets, solar system. Not open for credit to students who have taken ASTR 102 or ASTR 301; not open to upper-division students majoring in physical sciences or engineering. Offered: AWSpS.
<BR>Instructor Course Description:
<A HREF="/students/icd/S/astro/101anamunn.html"><I>Ana M. Larson</I></A>
<A HREF="/students/icd/S/astro/101ojf.html"><I>Oliver J. Fraser</I></A>
<A HREF="/students/icd/S/astro/101paulas.html"><I>Paula Szkody</I></A>

<P><B><A NAME="astr102">ASTR 102 </A> Introduction to Astronomy (5) NW, QSR </B><BR>Emphasis on mathematical and physical comprehension of nature, the sun, stars, galaxies, and cosmology. Designed for students who have had algebra and trigonometry and high school or introductory-level college physics. Cannot be taken for credit in combination with ASTR 101 or ASTR 301. Offered: A.
<BR>Instructor Course Description:
<A HREF="/students/icd/S/astro/102balick.html"><I>Bruce Balick</I></A>

<P><B><A NAME="astr105">ASTR 105 </A> Exploring the Moon (5) NW </B><I> Smith </I><BR>Examines the questions why did we go to the moon, what did we learn, and why do we want to go back. Offered: W.

`
)

func TestExtractClassDescriptions(t *testing.T) {
	classDescriptionIndex := ClassDescriptionIndex(sampleClassDescriptionIndex)
	for k, v := range classDescriptionIndex.Extract() {
		fmt.Println(k, "==============================")
		fmt.Println(v)
	}
}
