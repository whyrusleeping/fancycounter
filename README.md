# Fancy Counter

Do you need to count occurences of certain numbers within large lists of
numbers to find out which of those numbers occur the most often?

Do you need it to be fast?

Well I've got a data structure for you! 

Introducing... The Fancy Counter!

Its a stack of roaring bitmaps that effectively comprise a compressed N bit integer for the number of each item in the set (up to the maximum value configured when setting the thing up)

This is used in Bluesky's Discover algorithm for suggesting who to follow based on who the people you are following follow.

## License
MIT
