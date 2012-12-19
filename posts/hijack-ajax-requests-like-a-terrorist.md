--- 
id: 446
author: Daniel Huckstep
title: Hijack AJAX Requests Like A Terrorist
category: programming
description: I show you how to hijack AJAX requests.
published: true
publishedon: 20 Feb 2010 18:00 MST
slugs: 
- hijack-ajax-requests-like-a-terrorist
tags: 
- ajax
- web
- javascript
- jquery
images: 
  intercept_ajax: 
    small: http://cdn.verboselogging.com/transloadit/small/5f/4d5b14ee10f3dfc67a71d0dd46ee42/intercept-ajax.png
    medium: http://cdn.verboselogging.com/transloadit/medium/0a/a4014d183da1be296b81a31c2e4842/intercept-ajax.png
    large: http://cdn.verboselogging.com/transloadit/large/88/659d05da0b90ee51e3704a2d71d675/intercept-ajax.png
    original: http://cdn.verboselogging.com/transloadit/original/43/470c6b7fa8f6fdc599545dd7fd45dd/intercept-ajax.png
---
<p><span class="caps">AJAX</span> requests are a grand thing. They let you request things from your server without refreshing the page. Now, if you are trying to proxy a page, you can rewrite all the links in the page to point back through your proxy, but <span class="caps">AJAX</span> requests are another thing.</p>
<h5>Oh wait no they&#8217;re not!</h5>
<p>You can&#8217;t rewrite them when you proxy the page (by proxy, I mean you request my page with a <span class="caps">URL</span> param to another page, and I pull in that page, do some stuff, and serve it to you), but you still want the <span class="caps">AJAX</span> to go through your proxy, since otherwise it won&#8217;t work.</p>
<p>Luckily there&#8217;s a solution!</p>
<p>No matter what framework you use, <a href="http://jquery.com/">jQuery</a>, <a href="http://prototypejs.org/">Prototype</a>, whatever, they all go through the <a href="http://en.wikipedia.org/wiki/XMLHttpRequest">XMLHttpRequest</a> interface. That is unless you are rockin&#8217; IE6, in which case they use an <code>ActiveXObject</code>. I don&#8217;t deal with that, although I&#8217;m sure you can do something similar with it.</p>
<p>Anyway.</p>
<p>So you have this <code>XMLHttpRequest</code> thing, and as an example, in the jQuery code they do this:</p>
<pre>new window.XMLHttpRequest();</pre>
<p>See that <code>new</code> in there? They are creating a new <em>object</em> (for varying definitions of <em>object</em>). But whatever, this means we can use the magic of <strong>prototype</strong>. There&#8217;s <a href="http://www.howtocreate.co.uk/tutorials/javascript/objects">a bunch</a> <a href="http://www.packtpub.com/article/using-prototype-property-in-javascript">of stuff</a> <a href="http://stackoverflow.com/questions/572897/how-does-javascript-prototype-work">out there</a> on prototype, so I won&#8217;t cover it, but let&#8217;s get some code.</p>
<script type="text/javascript" src="http://gist.github.com/309973.js?file=ajaxIntercept.js"></script><p>And it works like this:</p>
<p><figure><a href="http://cdn.verboselogging.com/transloadit/original/43/470c6b7fa8f6fdc599545dd7fd45dd/intercept-ajax.png"><img src="http://cdn.verboselogging.com/transloadit/large/88/659d05da0b90ee51e3704a2d71d675/intercept-ajax.png" class=" large" alt="" /></a></figure></p>
<p>So check this out. First, we define an anonymous function that we call immediately:</p>
<pre>(function() {
})();</pre>
<p>The reason we need to do this is so that we can have a reference to the original <code>open</code> method without having to have other weird things kicking around just for that. So we call the method with the original <code>open</code> method as the only parameter:</p>
<pre>(function(open) {
})(XMLHttpRequest.open);</pre>
<p>Then with the prototype method, we redefine the <code>code</code> method on all XMLHttpRequest objects:</p>
<pre>XMLHttpRequest.prototype.open = function(method, url, async, user, pass) { ... }</pre>
<p>While keeping the original method around so we can intercept calls to it.</p>
<pre>// Do some magic
open.call(...);</pre>
<p>Put it all together and you get the <span class="caps">AJAX</span> interception code.</p>
<script type="text/javascript" src="http://gist.github.com/309973.js?file=ajaxIntercept.js"></script><p>Simply replace the <code>// Do some magic</code> comment with your code to rewrite the <span class="caps">URL</span>, or do whatever with the request. Now when you proxy the request, just prepend a script tag to the <code>head</code> element (make it the first element inside the <code>head</code> tag) so it gets loaded before any other of the scripts on the page.</p>