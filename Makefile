##
# Project Title
#
# @file
# @version 0.1



# end

run_test:
	docker build -t packtraktest1 . -f test/Dockerfile.1
	docker run packtraktest1
