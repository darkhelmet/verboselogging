--- 
id: 439
author: Daniel Huckstep
title: Proper Rack Middleware Ordering
category: programming
description: I show the importance of properly ordering rack middleware.
published: true
publishedon: 20 Jan 2010 08:00 MST
slugs: 
- proper-rack-middleware-ordering
tags: 
- sinatra
- rack
- middleware
- ruby
images: 
  bad_middleware: 
    small: http://cdn.verboselogging.com/transloadit/small/9e/a982a8c20c70d6d3ed7c72e37bc45d/bad-middleware.png
    large: http://cdn.verboselogging.com/transloadit/large/29/9403ab5fa6494e677422f299104e05/bad-middleware.png
    medium: http://cdn.verboselogging.com/transloadit/medium/42/7afc23044cc6435965b362f6eb9d8f/bad-middleware.png
    original: http://cdn.verboselogging.com/transloadit/original/14/59a6413b93dc74364caa0b0ae7038d/bad-middleware.png
  good_middleware: 
    small: http://cdn.verboselogging.com/transloadit/small/91/d8ce4472f3d886d54317a5b755a6f0/good-middleware.png
    large: http://cdn.verboselogging.com/transloadit/large/b5/0c9f9d808c6a451998ec502310a65b/good-middleware.png
    medium: http://cdn.verboselogging.com/transloadit/medium/d2/e6b7d47686e5f74195f02b3230421d/good-middleware.png
    original: http://cdn.verboselogging.com/transloadit/original/4b/2f4ad4b80f1bb254e8347c124a8b0f/good-middleware.png
---
<p>It occurred to me the other day, that I should take a look at the middleware I use on this blog. I don&#8217;t know what it was. My spidey senses just tingled.</p>
<p>Boy was I right. I totally had it backwards.</p>
<p>Rack middleware is a fantastic thing. It&#8217;s like a little encapsulated rack application that you can use to filter, process, or otherwise mess with responses. There is middleware to add <a href="http://github.com/rack/rack/blob/master/lib/rack/etag.rb">etags</a>, <a href="http://github.com/rtomayko/rack-cache">configure caching</a>, catch and log exceptions, deal with cookies, handle <a href="http://en.wikipedia.org/wiki/Single_sign-on"><span class="caps">SSO</span></a>, and pretty much anything else you can think of. Oh, and they work on any rack application; it is <em>rack</em> middleware after all. And in case you missed it, rails is a rack application. Create a new rails app and run</p>
<pre>% rake middleware</pre>
<p>You&#8217;ll see all the middleware that is included by default.</p>
<p>Anyway. The thing with rack middleware is that it runs in the order you specify them, top to bottom, and then by nature of how they work, they sort of rewind out.</p>
<p>Okay so <span class="caps">WTF</span> does that mean? A basic middleware looks kind of like this:</p>
<script type="text/javascript" src="http://gist.github.com/281637.js?file=etag.rb"></script><p>That&#8217;s the etag middleware. It adds an etag value to responses. The required parts are the initialize method taking the application (which is a rails app, sinatra app, whatever), and the call method, taking an environment. Initialize sets things up, and call is what happens when a request comes in. The whole idea is you do:</p>
<pre>@app.call(env)</pre>
<p>In <em>your</em> call method, where @app could be another middleware, or the actual application, but regardless it eventually gets all the way down to the real application. As the methods return, it comes back up with a response body, headers, and status code. In the etag example, @app.call(env) is done immediately and the results processed; the etag value is set in the headers.</p>
<p>So let&#8217;s think about this for a second. Image you have some setup like this:</p>
<pre>use Rack::Etag
use Rack::ResponseTimeInjector
use Rack::Hoptoad</pre>
<p>Does that really make sense? When you <em>use</em> middleware, you&#8217;re telling your framework or whatever to <em>append</em> that middleware to the chain. So request comes in, goes through middleware, then hits your application.</p>
<p>In this case:</p>
<ol>
	<li>We come into Etag&#8230;</li>
	<li>&#8230;which calls the ResponseTimeInjector middleware&#8230;</li>
	<li>&#8230;which calls the Hoptoad middleware to catch exceptions&#8230;</li>
	<li>&#8230;which calls your application&#8230;</li>
	<li>&#8230;which returns to the Hoptoad middleware&#8230;</li>
	<li>&#8230;which returns to the ResponseTimeInjector middleware&#8230;</li>
	<li>&#8230;which inserts the response time into the body&#8230;</li>
	<li>&#8230;which then returns (with the modified body) to the Etag middleware&#8230;</li>
	<li>&#8230;which calculates the etag value and puts it in the headers&#8230;</li>
	<li>&#8230;which returns and lets rack send the response back.</li>
</ol>
<p>Whew! Lots of steps there, but this might make more sense:</p>
<p><figure><a href="http://cdn.verboselogging.com/transloadit/original/14/59a6413b93dc74364caa0b0ae7038d/bad-middleware.png"><img src="http://cdn.verboselogging.com/transloadit/large/29/9403ab5fa6494e677422f299104e05/bad-middleware.png" class=" large" alt="" /></a></figure></p>
<p>Okay so what&#8217;s the problem? The etag is calculated <em>after</em> the response time is injected, so that&#8217;s fine (imagine if the etag middleware was at the bottom). What about poor Hoptoad? What if there is an exception thrown in the ResponseTimeInjector or Etag middleware? Hoptoad isn&#8217;t going to catch it! The Hoptoad middleware doesn&#8217;t modify anything in the response, so it needs to be up higher; it needs to be first.</p>
<pre>use Rack::Hoptoad
use Rack::Etag
use Rack::ResponseTimeInjector</pre>
<p>Diagram time:</p>
<p><figure><a href="http://cdn.verboselogging.com/transloadit/original/4b/2f4ad4b80f1bb254e8347c124a8b0f/good-middleware.png"><img src="http://cdn.verboselogging.com/transloadit/large/b5/0c9f9d808c6a451998ec502310a65b/good-middleware.png" class=" large" alt="" /></a></figure></p>
<p>That&#8217;s better! This is basically the problem I had, except worse. I don&#8217;t know what I was thinking, but my middleware was all out of order: <a href="http://github.com/darkhelmet/darkblog/blob/42483fa463c7891967a908d6792b27f4aea57d21/lib/middleware.rb">before</a> and <a href="http://github.com/darkhelmet/darkblog/blob/f19fecfd4b4cf453e9e46119a1e9aa6d95aa17f0/lib/middleware.rb">after</a>.</p>
<p>See that? It&#8217;s gross! I had Etag near the bottom, my exception logger was <em>all</em> the way at the bottom. The only one that was in a remotely right place was CanonicalHost.</p>
<p>The really terrible part about the original was that the body was being modified by 4 different middleware classes <em>after</em> the etag middleware runs and returns, hence the etag was wrong.</p>
<p>So hopefully if you are a rack middleware nerd already, you probably knew this stuff and stopped reading a while ago, or you are laughing at me. Otherwise, you might consider thinking twice about how your middleware is organized. Maybe go take a look at your middleware stack anyway and see if anything can be optimized.</p>
<p>Now go use some rack middleware! Ba dum tiss! (See that, see what I did there? In code you <em>use</em> middleware, and at a higher level as a developer you use middleware, so &#8230; ah nevermind)</p>