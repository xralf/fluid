# FOO="x y z" BAR="abc def" make foo
# Hello, here are the arguments: x y z abc def

.PHONY: foo

all:
	ls -lsa
	date
	pwd

foo:
	@echo "Hello, here are the arguments:" $(FOO) $(BAR)
