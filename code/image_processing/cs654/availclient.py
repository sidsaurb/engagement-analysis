import urllib2
import urllib
import sys

fd = open(sys.argv[1], 'r')
l = fd.read()
fd.close()

#print len(l)

values = {}
values['file'] = l
data = urllib.urlencode(values)
req = urllib2.Request('http://localhost:9999', data)
resp = urllib2.urlopen(req)
content = resp.read()
print content
