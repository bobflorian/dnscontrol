# DNSControl

DNSControl is a system for scripting your DNS configuration independently of where your domains are registered or hosted.
At Stack Overflow, we use this system to manage hundreds of domains and subdomains across multiple registrars and DNS providers.

## Installation

`go get github.com/StackExchange/dnscontrol`

or get prebuilt binaries from our github page.

### Configuration

Configuration is provided via a javascript file interpreted by dnscontrol. This allows more flexibility than a simple config file, as you can define variables and reuse sections multiple times.

Here is an example **dnsconfig.js**:

```
//domains registered with name.com
var REG_NAMECOM = NewRegistrar("nc","NAMEDOTCOM");

//dns handled by cloudflare
var DSP_CLOUDFLARE = NewDSP("CF","CLOUDFLAREAPI");

var addr = IP("1.2.3.4");
var addr2 = IP + 2;

D( "example.com", REG_NAMECOM, //name and registrar
   DSP_CLOUDFLARE,             //dns provider
   A("@", addr),
   A("www",addr),
   A("blog",addr2, TTL(300)),
   CNAME("mail","foo.com.")
)
```

Full configuration documentation [here](docs/js.md).

An additional config file is required with account details and credentials. **providers.json:**

```
{
   "CF": { //name must match declared name in dnsconfig.js
      "apikey": "654e27bbe212654e27bbe212",
      "apiuser": "my.email@domain.tld"
   },
   "nc":{
      "apikey": "654e27bbe212654e27bbe212",
      "apiuser": "myusername"
   }
}
```
### Running

**preview needed corrections:**

`dnscontrol preview`

**perform corrections:**

`dnscontrol push`


 