--- 
id: 492
author: Daniel Huckstep
title: Proc, Block, and Two Smoking Lambdas
category: programming
description: The differences between the closure syntaxes in Ruby 1.9.
published: true
publishedon: 20 Sep 2011 10:00 MDT
slugs: 
- proc-block-and-two-smoking-lambdas
tags: 
- ruby
- closure
- syntax
images: 
  vinnie: 
    small: http://cdn.verboselogging.com/transloadit/small/db/8da51eef09f6549ed52a9444fa0201/vinnie.jpg
    medium: http://cdn.verboselogging.com/transloadit/medium/82/e37bbd0d09a85b24b1265f4a046160/vinnie.jpg
    original: http://cdn.verboselogging.com/transloadit/original/c0/eb484a3e3e543262eb884ec0ae692c/vinnie.jpg
    large: http://cdn.verboselogging.com/transloadit/large/3a/5b1ab613b31ffc526d548f8fe7ecff/vinnie.jpg
---
<p><figure><img src="http://cdn.verboselogging.com/transloadit/medium/82/e37bbd0d09a85b24b1265f4a046160/vinnie.jpg" class="fright bleft bbottom round medium" alt="" /></figure></p>
<p>Ruby 1.9 has 4 different ways to deal with closures.</p>
<p><strong>Cue music</strong></p>
<h2>Proc</h2>
<p>Procs are the weird ones of the bunch. Technically, all of these things I&#8217;m going to describe are Procs. By that I mean, if you check the <code>class</code>, it&#8217;s a <code>Proc</code>.</p>
<p>A <code>Proc</code> is made by using <code>Proc.new</code> and passing a block, <code>Proc.new { |x| x }</code>, or by using the <code>proc</code> keyword, <code>proc { |x| x }</code>.</p>
<p>A <code>return</code> from inside exits completely out of the method enclosing the <code>Proc</code>.</p>
<p>A <code>Proc</code> doesn&#8217;t care about the arguments passed. If you define a <code>Proc</code> with two parameters, and you pass only 1, or possibly 3, it keeps on trucking. In the case of 1 argument, the second parameter will have the value <code>nil</code>. If you pass extra arguments, they will be ignored and lost.</p>
<h2>Block</h2>
<p>Blocks are when you pass an anonymous closure to a method:</p>
<pre>def my_method
  my_other_method(1) do |x, y|
    return x + y
  end
end</pre>
<p>They work exactly like a <code>Proc</code>. It wouldn&#8217;t matter how many arguments <code>my_other_method</code> called <code>yield</code> with, the block would execute just fine.<sup class="footnote" id="fnr1"><a href="#fn1">1</a></sup> The <code>return</code> will also return out of <code>my_method</code>.</p>
<h2>Lambda</h2>
<p>A <code>lambda</code> is probably what you deal with most of time. You make them with the <code>lambda</code> keyword: <code>f = lambda { |x| x + 1 }</code>. They are a bit different.</p>
<p>Unlike a <code>Proc</code>, using <code>return</code> in a <code>lambda</code> will simply return from the <code>lambda</code>, pretty much like you&#8217;d expect.</p>
<p>Also unlike a <code>Proc</code>, <code>lambda</code> likes to whine if you pass an incorrect number of arguments. It will blow up with an <code>ArgumentError</code>.</p>
<h2>Stabby</h2>
<p>The stabby is new in Ruby 1.9, and is just syntactic sugar for <code>lambda</code>. These are equivalent:</p>
<pre>f = lambda { |x| x + 1 }</pre>
<pre>f2 = -&gt;(x) { x + 1 }</pre>
<h2>What&#8217;s all this then?</h2>
<p>So anyway I wrote some specs, and here they are (or rather their output). If you want to check out the actual specs, or run them for yourself, head on over to <a href="https://github.com/darkhelmet/proc-block">Github</a>.</p>
<script src="https://gist.github.com/1224675.js?file=out.txt"></script><p class="footnote" id="fn1"><a href="#fnr1"><sup>1</sup></a> The addition won&#8217;t work too well, but hey.</p>
