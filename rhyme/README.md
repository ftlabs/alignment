# experimenting with matching words by rhyme

stole a big file of word syllables from http://svn.code.sf.net/p/cmusphinx/code/trunk/cmudict/cmudict-0.7b

documentation: http://www.speech.cs.cmu.edu/cgi-bin/cmudict

Its entries are particularly useful for speech recognition and synthesis, as it has mappings from words to their pronunciations in the ARPAbet phoneme set, a standard for English pronunciation. The current phoneme set contains 39 phonemes, vowels carry a lexical stress marker:

* 0    — No stress
* 1    — Primary stress
* 2    — Secondary stress

## iambic pentameter

from [wikipedia](https://en.wikipedia.org/wiki/Iambic_pentameter)

An iambic foot is an unstressed syllable followed by a stressed syllable. The rhythm can be written as:

da DUM

The da-DUM of a human heartbeat is the most common example of this rhythm.

A standard line of iambic pentameter is five iambic feet in a row:

da DUM da DUM da DUM da DUM da DUM

Straightforward examples of this rhythm can be heard in the opening line of William Shakespeare's Sonnet 12:

When I do count the clock that tells the time

and in John Keats' To Autumn:[1]

To swell the gourd, and plump the hazel shells

## unrecognised words

Whilst the base CMUDict is huge, english as wot is wrote is huger, so lots of words escape the poetry filter. 

At the foot of most of the web pages there is a list of 'unrecognised words'. These can be duly recognised by being added to cmudict-0.7b_my_additions, which extends the base CMUDict with extra words and features. The comments in that file should explain all.
