# DNSControl

DNSControl is a system for maintaining DNS zones.  It has two parts:
a domain specific language (DSL) for describing DNS zones plus
software that processes the DSL and pushes the resulting zones to
DNS providers such as Route53, CloudFlare, and Gandi.  It can talk
to Microsoft ActiveDirectory and it generates the most beautiful
BIND zone files ever.  It run anywhere Go runs (Linux, macOS,
Windows).

At Stack Overflow, we use this system to manage hundreds of domains
and subdomains across multiple registrars and DNS providers.

You can think of it as a DNS compiler.  The configuration files are
written in a DSL that looks a lot like JavaScript.  It is compiled
to an intermediate representation (IR).  Compiler back-ends use the
IR to update your DNS zones on services such as Route53, CloudFlare,
and Gandi, or systems such as BIND and ActiveDirectory.

# Benefits

* Editing zone files is error-prone.  Clicking buttons on a web
page is irreproducible.
* Switching DNS providers becomes a no-brainer.  The DNSControl
language is vendor-agnostic.  If you use it to maintain your DNS
zone records, you can switch between DNS providers easily. In fact,
DNSControl will upload your DNS records to multiple providers, which
means you can test one while switching to another. We've switched
providers 3 times in three years and we've never lost a DNS record.
* Adopt CI/CD principles to DNS!  At StackOverflow we maintain our
DNSControl configurations in Git and use our CI system to roll out
changes.  Keeping DNS information in a VCS means we have full
history.  Using CI enables us to include unit-tests and system-tests.
Remember when you forgot to include a "." at the end of an MX record?
We haven't had that problem since we included a test to make sure
Tom doesn't make that mistake... again.
* Variables safe time!  Assign an IP address to a constant and use
the variable name throughout the file. Need to change the IP address
globally? Just change the variable and "recompile."
* Macros!  Define your SPF records, MX records, or other repeated
data once and re-use them for all domains.
* Control CloudFlare from a single location.  Enable/disable
Cloudflare proxying (the "orange cloud" button) directly from your
DNSControl files.
* Keep similar domains in sync with transforms and other features.
If one domain is supposed to be the same
* It is extendable!  All the DNS providers are written as plugins.
Writing new plugins is very easy.

# Installation

`go get github.com/StackExchange/DNSControl`

or get prebuilt binaries from our github page.

# A simple example

The DNSControl configuration language is (essentially) JavaScript.
However you don't need to know JavaScript except for defining
variables and domains.

A simple example **dnsconfig.js**:

```
var REG_NAMECOM = NewRegistrar("nc","NAMEDOTCOM");
// This defines that domains marked REG_NAMECOM will talk to their
// registrar using the NAMEDOTCOM plugin and will get their credentials
// from the "nc" stanza in creds.json.

var DSP_CLOUDFLARE = NewDSP("CF","CLOUDFLAREAPI");
// This defines that domains marked DSP_CLOUDFLARE will talk to
// their DNS Service Provider using the CLOUDFLAREAPI plugin and
// will get their credentials from the "CF" stanza in creds.json.
// You have to define a Registrar and a DSP even if the same plugin
// and credientials are used for both (because usually they aren't).

// Define the IP addresses we'll use. We can also just specify
// them as quoted strings.
var addr = IP("1.2.3.4");  // IP() lets you perform math on an IP address.
var addr2 = IP + 2;

D("example.com", REG_NAMECOM,  // DNS zone name and DNS REGistrar
   DSP_CLOUDFLARE,             // DNS Service Provider
   A("@", addr),
   A("www", addr),
   A("blog",addr2, TTL(300)),  // TTL() specifies a special TTL.
   CNAME("mail","foo.com."),
   A("quotes", '1.2.3.10'),    // You can also just use quotes
   A("work", "1.2.3.10")       // or double-quotes.
)
```

Full configuration documentation [here](docs/js.md).

Credentials are stored in a separate file called **creds.json**.

```
{
  "CF": {
    "apikey": "654e27bbe212654e27bbe212",
    "apiuser": "my.email@domain.tld"
  },
  "nc": {
    "apikey": "654e27bbe212654e27bbe212",
    "apiuser": "myusername"
  }
}
```

NOTE: The names in the file must match the declared name in dnsconfig.js exactly.


### Running

Compile but don't actually make the changes:

`dnscontrol preview`

Compile and push changes to the providers:

`dnscontrol push`

Only process certain domains:

`dnscontrol -domains one.com,two.com push`


### A big, complex, example:


```
// Registrar(s):
var REG_NAMECOM = NewRegistrar("nc","NAMEDOTCOM");
// DNS Providers(s):
var DSP_NAMECOM = NewDSP("nc","NAMEDOTCOM");
var DSP_CLOUDFLARE = NewDSP("CF","CLOUDFLAREAPI");
// NOTE: The creditials for these APIs are found in
// creds.json.

# CONSTANTS:

// Our IP address space starts at MAIN.
var MAIN = IP('1.2.3.0')
// var MAIN = IP('2.3.4.0')  // Use this when failover'ed to DR site.
var SSHSERVER  = MAIN+10
var MAINWEB    = MAIN+11
var DOCWEB     = MAIN+12
var SUPPORTWEB = MAIN+13
var PARKED     = MAIN+14
var PARKED_IPV6 = '2607:f440::1234:1234'

// These MX records are for any domain that uses Google Apps:
var GOOGLE_APPS_DOMAIN_MX = [
  MX('@', 1, 'aspmx.l.google.com.'),
  MX('@', 5, 'alt1.aspmx.l.google.com.'),
  MX('@', 5, 'alt2.aspmx.l.google.com.'),
  MX('@', 10, 'aspmx2.googlemail.com.'),
  MX('@', 10, 'aspmx3.googlemail.com.'),
]

// Cloudflare Macros:
var CF_PROXY_OFF = {'cloudflare_proxy': 'off'};     // Default/off.
var CF_PROXY_ON = {'cloudflare_proxy': 'on'};       // Sites safe to proxy.
// When Cloudflare is down, uncomment this lines and pray:
// var CF_PROXY_ON = CF_PROXY_OFF

// Sneaky trick so other lines can end with ",".
// This makes diffs shorter.
var END = {}

// DOMAINS:

D("the-cloud-book.com", REG_NAMECOM, DSP_CLOUDFLARE,
    GOOGLE_APPS_DOMAIN_MX,
    A('@', MAINWEB, CF_PROXY_ON),
    A('docs', DOCWEB, CF_PROXY_ON),
    A('support', SUPPORTWEB, CF_PROXY_OFF),
    A('ssh', SSHSERVER),
    AAAA('@', MAINWEB_IPV6, CF_PROXY_ON),
    CNAME('www', '@', CF_PROXY_ON),
END)

D("example.com", REG_NAMECOM, DSP_CLOUDFLARE,
    A('@', MAINWEB, CF_PROXY_ON),
    AAAA('@', MAINWEB_IPV6, CF_PROXY_ON),
    CNAME('www', '@', CF_PROXY_ON),
END)

// This domain is the same as another domain, with some changes.
// Now if we change that domain, this one is updated automatically.
// Being able to generate one domain based on another eliminates
// the problem where someone updates one zone and forgets to update
// the other. You generally don't want to use this, but when there
// is no other choice, it reduces a lot of pain.
var TF1 = [
    {low: "1.2.3.11", high: "1.2.3.11", newBase: '99.99.99.99' },
]
D("derived-domain.com", REG_NAMECOM, DSP_NAMECOM,
    IMPORT_TRANSFORM(TF1, 'example.com')
)

// Parked domains all have the same settings, so why
// not use a loop?

var a = [
  'parkeddomain1.com',
  'parkeddomain2.com',
  'theotherparkeddomain.com',
  'fakeparkdomain.com',
  'sampleparkeddomain.com',
];
for (index in a) {
  D(a[index], REG_NAMECOM, DSP_GANDI,
    A('@', PARKED),
    A('*', PARKED),
    AAAA('@', PARKED_IPV6),
    TXT('@', 'v=spf1 -all'),
  END)
}

// If you don't like using a loop, you can specify a macro and use
// it for many domains.
var FOR_SALE = [
  A('@', '1.2.3.4'),
  TXT("@", "v=spf1 mx -all")  // No email from anywhere but MX.
];
D('forsale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE)
D('fivesale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE)
D('sixsale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE,
  AAAA("*", '2607:f440::1234:1234'),  // This domain is a little differnet.
END)
D('sevensale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE)
D('eightsale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE)
D('ninesale.com', REG_NAMECOM, DSP_NAMECOM, FOR_SALE)
```
