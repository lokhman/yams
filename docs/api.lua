local os = require("os")
local string = require("string")
local yams = require("yams")

-- Header --
yams.write([=[<!doctype html>
<html>
<head>
  <title>YAMS Docs</title>
</head>
<body>
  <h1><img src="/yams.png" style="width: 64px; vertical-align: middle;"> YAMS Docs</h1>]=])

-- TOC --
yams.write([=[
  <h2>Table of Contents</h2>
  <ul>
    <li>
      <strong>local yams = require("yams")</strong>
      <ul>
        <li><a href="#yams.routeid">yams.routeid</a></li>
        <li><a href="#yams.method">yams.method</a></li>
        <li><a href="#yams.scheme">yams.scheme</a></li>
        <li><a href="#yams.host">yams.host</a></li>
        <li><a href="#yams.uri">yams.uri</a></li>
        <li><a href="#yams.ip">yams.ip</a></li>
        <li><a href="#yams.sessionid">yams.sessionid</a></li>
        <li><a href="#yams.form">yams.form</a></li>
        <li><a href="#yams.path">yams.path</a></li>
        <li><a href="#yams.headers">yams.headers</a></li>
        <li><a href="#yams.query">yams.query</a></li>
        <li><a href="#yams.cookies">yams.cookies</a></li>
        <li><a href="#yams.setstatus">yams.setstatus(code)</a></li>
        <li><a href="#yams.getheader">yams.getheader(name)</a></li>
        <li><a href="#yams.setheader">yams.setheader(name, value [, value, ...])</a></li>
        <li><a href="#yams.setcookie">yams.setcookie(name, value [, expires [, path [, maxage [, secure [, httponly]]]]])</a></li>
        <li><a href="#yams.parseform">yams.parseform([maxmemory])</a></li>
        <li><a href="#yams.getparam">yams.getparam(name)</a></li>
        <li><a href="#yams.getbody">yams.getbody()</a></li>
        <li><a href="#yams.asset">yams.asset(path)</a></li>
        <li><a href="#yams.sleep">yams.sleep(seconds)</a></li>
        <li><a href="#yams.write">yams.write(...)</a></li>
        <li><a href="#yams.getvar">yams.getvar(name [, islocal])</a></li>
        <li><a href="#yams.setvar">yams.setvar(name [, value [, islocal [, lifetime]]])</a></li>
        <li><a href="#yams.dump">yams.dump([withbody])</a></li>
        <li><a href="#yams.wbclean">yams.wbclean()</a></li>
        <li><a href="#yams.pass">yams.pass([url])</a></li>
        <li><a href="#yams.exit">yams.exit()</a></li>
      </ul>
    </li>
    <li>
      <strong>local asset = yams.asset("path/to/asset")</strong>
      <ul>
        <li><a href="#asset:getmimetype">asset:getmimetype()</a></li>
        <li><a href="#asset:getsize">asset:getsize()</a></li>
        <li><a href="#asset:template">asset:template(vars)</a></li>
      </ul>
    </li>
    <li>
      <strong>local json = require("json")</strong>
      <ul>
        <li><a href="#json.encode">json.encode(value)</a></li>
        <li><a href="#json.decode">json.decode(data)</a></li>
      </ul>
    </li>
    <li>
      <strong>local base64 = require("base64")</strong>
      <ul>
        <li><a href="#base64.encode">base64.encode(value)</a></li>
        <li><a href="#base64.decode">base64.decode(data)</a></li>
      </ul>
    </li>
  </ul>
]=])

-- Docs --
yams.write([=[
  <h2>Lua Script API</h2>
  <p>
    Syntatically and functionally is based on Lua 5.1.
    Look into Reference Manual at <a href="https://www.lua.org/manual/5.1/">https://www.lua.org/manual/5.1/</a>.
  </p>
  <ul>
    <li>
      <h3>local yams = require("yams")</h3>
      <p>Library <code>yams</code> is the base extension that is used for communication with YAMS proxy.</p>
      <ul>
        <li>
          <h4><a href="#yams.routeid" name="yams.routeid">yams.routeid</a></h4>
          <p>Returns request route identifier (UUID). Same as <code>X-YAMS-Route-Id</code> header in debug mode.</p>
          <pre>
            yams.write("YAMS Route: " .. yams.routeid)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.method" name="yams.method">yams.method</a></h4>
          <p>Returns request method (e.g. <em>GET</em>, <em>POST</em>, <em>DELETE</em>, etc).</p>
          <pre>
            yams.write("Request method: " .. yams.method)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.scheme" name="yams.scheme">yams.scheme</a></h4>
          <p>Returns request scheme (e.g. <em>http</em> or <em>https</em>).</p>
          <pre>
            yams.write("Request scheme: " .. yams.scheme)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.host" name="yams.host">yams.host</a></h4>
          <p>Returns request host including port if provided (e.g. <em>yams.brandwidth.com</em>, etc).</p>
          <pre>
            yams.write("Request host: " .. yams.host)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.uri" name="yams.uri">yams.uri</a></h4>
          <p>Returns request URI starting with forward slash (e.g. <em>/path/to/resource</em>, etc).</p>
          <pre>
            yams.write("Request URI: " .. yams.uri)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.ip" name="yams.ip">yams.ip</a></h4>
          <p>Returns client's IP address (e.g. <em>37.252.27.254</em>, etc).</p>
          <pre>
            yams.write("IP address: " .. yams.ip)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.sessionid" name="yams.sessionid">yams.sessionid</a></h4>
          <p>Returns current session identifier. Same as <code>X-YAMS-Session-Id</code> header in debug mode.</p>
          <pre>
            yams.write("YAMS Session: " .. yams.sessionid)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.form" name="yams.form">yams.form</a></h4>
          <p>Returns table with form values. Before using this field, body needs to be parsed with <a href="#yams.parseform">yams.parseform</a> function.</p>
          <pre>
            yams.write("Value of form parameter `param1`: ", yams.form.param1[1])
          </pre>
        </li>
        <li>
          <h4><a href="#yams.path" name="yams.path">yams.path</a></h4>
          <p>Returns table with request path parameters.</p>
          <pre>
            yams.write("Value of path parameter `param1`: ", yams.path.param1)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.headers" name="yams.headers">yams.headers</a></h4>
          <p>Returns table with request headers (every header may have multiple values).</p>
          <pre>
            yams.write("Value of request header `Content-Type`: ", yams.headers["Content-Type"][1])
          </pre>
        </li>
        <li>
          <h4><a href="#yams.query" name="yams.query">yams.query</a></h4>
          <p>Returns table with request query string parameters (every parameter may have multiple values).</p>
          <pre>
            yams.write("Value of query string parameter `param1`: ", yams.query.param1[1])
          </pre>
        </li>
        <li>
          <h4><a href="#yams.cookies" name="yams.cookies">yams.cookies</a></h4>
          <p>Returns table with request cookies.</p>
          <pre>
            yams.write("Value of cookie `cookie1`: ", yams.cookies.cookie1)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.setstatus" name="yams.setstatus">yams.setstatus(code)</a></h4>
          <p>Sets response status code as per <code>code</code> argument. Can be overridden.</p>
          <pre>
            yams.setstatus(403)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.getheader" name="yams.getheader">yams.getheader(name)</a></h4>
          <p>Returns the first header value by header <code>name</code> from the request, or empty string if header was not found.</p>
          <pre>
            yams.write("Value of request header `Content-Type`: ", yams.getheader("Content-Type"))
          </pre>
        </li>
        <li>
          <h4><a href="#yams.setheader" name="yams.setheader">yams.setheader(name, value [, value, ...])</a></h4>
          <p>Sets header <code>value</code> or multiple values by <code>name</code> in the response.</p>
          <pre>
            yams.setheader("Content-Type", "text/csv")
          </pre>
        </li>
        <li>
          <h4><a href="#yams.setcookie" name="yams.setcookie">yams.setcookie(name, value [, expires [, path [, maxage [, secure [, httponly]]]]])</a></h4>
          <p>Sets cookie with <code>name</code> and <code>value</code> in the response. If <code>expires</code> is set as 0, the cookie will never expire; for all negative values, the cookie will be deleted.</p>
          <pre>
            yams.setcookie("user", "myuser", 3600, "/", 86400, false, true)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.parseform" name="yams.parseform">yams.parseform([maxmemory])</a></h4>
          <p>
            Parses the form or mulipart form body info a <a href="#yams.form">form</a> table.
            This function should be called before getting form values from <a href="#yams.form">form</a> table.
          </p>
          <pre>
            yams.parseform(5 * 2 ^ 20)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.getparam" name="yams.getparam">yams.getparam(name)</a></h4>
          <p>Returns query string or path or form parameter by <code>name</code>, or <code>nil</code> if parameter was not found.</p>
          <pre>
            yams.write("Value of any request parameter `param1`: ", yams.getparam("param1"))
          </pre>
        </li>
        <li>
          <h4><a href="#yams.getbody" name="yams.getbody">yams.getbody()</a></h4>
          <p>
            Returns body from the request or <code>nil</code> if body does not exist.
            This function cannot be used along with <a href="#yams.parseform">yams.parseform</a> function.
          </p>
          <pre>
            local data = json.decode(yams.getbody())
          </pre>
        </li>
        <li>
          <h4><a href="#yams.asset" name="yams.asset">yams.asset(path)</a></h4>
          <p>Returns <em>asset</em> by <code>path</code>, or <code>nil</code> if asset was not found.</p>
          <pre>
            local asset = yams.asset("path/to/asset")
            yams.setheader("Content-Type", asset:getmimetype())
            yams.write(asset)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.sleep" name="yams.sleep">yams.sleep(seconds)</a></h4>
          <p>Pauses script execution for the given number of <code>seconds</code>. Sleep duration cannot be higher than the defined route timeout.</p>
          <pre>
            yams.sleep(15)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.write" name="yams.write">yams.write(...)</a></h4>
          <p>
            Receives any number of arguments, and writes their values to the response output buffer, using the <code>tostring</code> function to convert them to strings.
            <em>asset</em> variables will be written directly to the response without memory buffering.
          </p>
          <pre>
            yams.write("Hello world!")
          </pre>
        </li>
        <li>
          <h4><a href="#yams.getvar" name="yams.getvar">yams.getvar(name [, islocal])</a></h4>
          <p>
            Returns variable value by <code>name</code> from the internal storage, or <code>nil</code> if variable was not found.
            If <code>islocal</code> is set to <code>true</code> (default <code>false</code>), variable will be read from the session storage by <a href="#yams.sessionid">yams.sessionid</a>.
          </p>
          <pre>
            yams.write("Value of session storage variable `var1`: ", yams.getvar("var1", true))
          </pre>
        </li>
        <li>
          <h4><a href="#yams.setvar" name="yams.setvar">yams.setvar(name [, value [, islocal [, lifetime]]])</a></h4>
          <p>
            Saves variable with the given <code>name</code> and <code>value</code> to the internal storage. Passing <code>value</code> as <code>nil</code> will remove variable from the storage.
            If <code>islocal</code> is set to <code>true</code> (default <code>false</code>), variable will be saved to the session storage by <a href="#yams.sessionid">yams.sessionid</a>.
            <code>lifetime</code> by default equals to the defined profile setting but can be overridden if smaller value is given.
          </p>
          <pre>
            yams.setvar("var1", "abc123", true, 3600)
          </pre>
        </li>
        <li>
          <h4><a href="#yams.dump" name="yams.dump">yams.dump([withbody])</a></h4>
          <p>Dumps the request with or without body (default <code>false</code>) to the response, then stops execution. This function clears response output buffer.</p>
          <pre>
            if yams.getparam("debug") then
              yams.dump(true)
            end
            yams.write("Hello world!")
          </pre>
        </li>
        <li>
          <h4><a href="#yams.wbclean" name="yams.wbclean">yams.wbclean()</a></h4>
          <p>Clears response output buffer (i.e. ignores all data written with <a href="#yams.write">yams.write</a>).</p>
          <pre>
            yams.write("This text will not be written.")
            yams.wbclean()
            yams.write("Hello world!")
          </pre>
        </li>
        <li>
          <h4><a href="#yams.pass" name="yams.pass">yams.pass([url])</a></h4>
          <p>Passes request to the backend if configured in the profile settings, otherwise requires <code>url</code> argument defined. This function stops script execution.</p>
          <pre>
            yams.pass("https://www.example.com")
          </pre>
        </li>
        <li>
          <h4><a href="#yams.exit" name="yams.exit">yams.exit()</a></h4>
          <p>Stops script execution, similar to <code>do return end</code>.</p>
          <pre>
            if yams.getparam("exit") then
              yams.exit()
            end
            yams.write("Hello world!")
          </pre>
        </li>
      </ul>
    </li>
  </ul>
  <ul>
    <li>
      <h3>local asset = yams.asset("path/to/asset")</h3>
      <p>User defined type <code>asset</code> which is returned from <a href="#yams.asset">yams.asset</a> function.</p>
      <ul>
        <li>
          <h4><a href="#asset:getmimetype" name="asset:getmimetype">asset:getmimetype()</a></h4>
          <p>Returns asset MIME-Type.</p>
          <pre>
            yams.write("Asset MIME-Type: ", asset:getmimetype())
          </pre>
        </li>
        <li>
          <h4><a href="#asset:getsize" name="asset:getsize">asset:getsize()</a></h4>
          <p>Returns asset size in bytes.</p>
          <pre>
            yams.write("Asset size: ", asset:getsize())
          </pre>
        </li>
        <li>
          <h4><a href="#asset:template" name="asset:template">asset:template(vars)</a></h4>
          <p>Returns compiled template from asset and given variables using Go <em>text/template</em> <a href="https://golang.org/src/text/template/doc.go" target="_blank">syntax</a>.</p>
          <pre>
            yams.write(yams.asset("path/to/asset.html"):template({ip=yams.ip}))
          </pre>
        </li>
      </ul>
    </li>
  </ul>
  <ul>
    <li>
      <h3>local json = require("json")</h3>
      <p>Library <code>json</code> is the extension that is used to manipulate JSON structures.</p>
      <ul>
        <li>
          <h4><a href="#json.encode" name="json.encode">json.encode(value)</a></h4>
          <p>Returns JSON string from the given <code>value</code>.</p>
          <pre>
            yams.write("Request headers: ", json.encode(yams.headers))
          </pre>
        </li>
        <li>
          <h4><a href="#json.decode" name="json.decode">json.decode(data)</a></h4>
          <p>Returns parsed into Lua structures JSON <code>data</code>.</p>
          <pre>
            local jsonbody = json.decode(yams.getbody())
            yams.write(jsonbody.prop1)
          </pre>
        </li>
      </ul>
    </li>
  </ul>
  <ul>
    <li>
      <h3>local base64 = require("base64")</h3>
      <p>Library <code>base64</code> is the extension that is used to encode and decode strings to and from Base64.</p>
      <ul>
        <li>
          <h4><a href="#base64.encode" name="base64.encode">base64.encode(value)</a></h4>
          <p>Returns Base64 encoded string from the given <code>value</code>.</p>
          <pre>
            yams.write("Encoded string: ", base64.encode(yams.getparam("param1")))
          </pre>
        </li>
        <li>
          <h4><a href="#base64.decode" name="base64.decode">base64.decode(data)</a></h4>
          <p>Returns decoded string from <code>data</code> in Base64.</p>
          <pre>
            yams.write("Decoded string: ", base64.decode(yams.getparam("param1")))
          </pre>
        </li>
      </ul>
    </li>
  </ul>]=])

yams.write(string.format([=[
  <hr>
  <address>
    Written by <a href="https://github.com/lokhman">Alex Lokhman II</a>.
    Generated with YAMS at %s
  </address>
]=], os.date()))

-- Footer --
yams.write([=[
  <script>
    Array.from(document.querySelectorAll('pre')).forEach(function(node) {
      node.textContent = node.textContent.replace(/^\s{10}|\s\s*$/gm, '')
    })
  </script>
</body>
</html>]=])
