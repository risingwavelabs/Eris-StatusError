- [x] update unittests 
- [x] additional values `map[string]any`
- [x] Additional status code field
- [x] GH: Run unit tests on each commit
- [ ] Update Readme 
- [x] Translation between gRPC err codes and our codes
- [ ] Read code docs `make docs`
- [ ] Read godoc docs
- [x] change API see below 

tests: 
test if map is printed 
test if custom serializer works 
If you cannot serialize it it should display something 
Do we ever use the print stuff? Do I test that or do I just test json?

API:
eris.New(message).WithCode(code).WithProperty(k,v)
eris.Newf(format, args...).WithCode(code).WithProperty(k,v)
eris.Wrap(err, message).WithCode(code).WithProperty(k,v)
eris.Wrapf(err, format, args...).WithCode(code).WithProperty(k,v)

