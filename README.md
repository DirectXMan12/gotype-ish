Gotype-ish
==========

This is a modified version of the standard gotype command.  It's modified
to actually check uncompiled sources when compiled versions aren't
available.

It's relatively quick, but for large projects should probably be run
asynchrously.  For example, with
[Neomake](https://github.com/neomake/neomake):

```vim
let g:neomake_go_gotype2_maker = {
    \ 'exe': '/path/to/gotype-ish',
    \ 'errorformat': '[%t] %f:%l:%c: %m',
    \ 'mapexpr': "printf('[E] %s', v:val)",
    \ }
let g:neomake_go_enabled_makers = ['gotype-ish']
```

To build, run `go build -o gotype-ish .`.

This code is licensed as the original gotype, found in LICENSE-GOTYPE.
