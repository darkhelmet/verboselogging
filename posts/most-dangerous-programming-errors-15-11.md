--- 
id: 463
author: Daniel Huckstep
title: Most Dangerous Programming Errors, 15-11
category: programming
description: I talk about 15-11 of the the Top 25 Most Dangerous Programming Errors.
published: true
publishedon: 31 May 2010 08:00 MDT
slugs: 
- most-dangerous-programming-errors-15-11
tags: 
- php
- ruby
- java
- dep
- aslr
- python
images: 
  spaghetti_os: 
    large: http://cdn.verboselogging.com/transloadit/large/6f/91ad08124eb58ce9dfdb08792cf782/spaghetti-os.jpg
    small: http://cdn.verboselogging.com/transloadit/small/9c/c8b6fe7527d60caccfcaf5de36035b/spaghetti-os.jpg
    original: http://cdn.verboselogging.com/transloadit/original/2b/7309d6abf824d82045acc0a2c0c6ee/spaghetti-os.jpg
    medium: http://cdn.verboselogging.com/transloadit/medium/ce/a09988eb66dfcb1dd268a1584f6efc/spaghetti-os.jpg
---
<p>It&#8217;s been a while, but I&#8217;ve been busy pwning n00bs at Modern Warfare 2 and Bad Company 2, and <a href="http://blog.darkhelmetlive.com/new-car-153">buying a car</a>, so life has been pretty busy as of late.</p>
<p>Have no fear though! I continue the look at the <a href="http://cwe.mitre.org/top25/index.html">Top 25 Most Dangerous Programming Errors</a> with numbers 15 to 11.</p>
<h2><a href="http://cwe.mitre.org/data/definitions/754.html">15. Improper Check for Unusual or Exceptional Conditions</a></h2>
<blockquote>
<p>When you <span class="caps">ASSUME</span> things, you make an <span class="caps">ASS</span> out of U and ME.</p>
</blockquote>
<p>This is all about assumptions. You assume something will work, you assume permissions will be set correctly, you assume there is a network connection.</p>
<p>In some cases, fine, make the assumptions and move on, but for the most part, you need to think these things through. Hackers like to abuse assumptions, and it can end badly.</p>
<p>The example they give is using <code>fgets</code> to read something, and then using <code>strcpy</code>.</p>
<script type='text/javascript' src='http://gist.github.com/418817.js?file=strcopy_fail.c'></script><p>Mad props on using <code>fgets</code> with the limit so you don&#8217;t overrun your buffer, but if an error occurs, the resultant string in <code>buf</code> might <strong>not</strong> be null terminated, in which case <code>strcpy</code> could run off the edge and explode. <code>strncpy</code> is better, and it should not be assumed that <code>fgets</code> will work flawlessly.</p>
<h3>Ways around it</h3>
<p>Don&#8217;t make assumptions.</p>
<p>Don&#8217;t assume that because you&#8217;re making a system call to something that <em>couldn&#8217;t possibly fail</em> you don&#8217;t have to check errors. There is always (okay, usually) something in the docs about what happens if somethings fails, so use that information and catch the case. File, console and network IO, file operations, and memory allocation can all fail, and while chances are they&#8217;ll work, there are those few times they don&#8217;t, and then your server gets owned!</p>
<h2><a href="http://cwe.mitre.org/data/definitions/129.html">14. Improper Validation of Array Index</a></h2>
<blockquote>
<p>There are only two hard things in Computer Science: cache invalidation, naming things, and off-by-one errors.</p>
</blockquote>
<p>While not the original quote, it&#8217;s pretty funny and illustrates the point. Improperly indexing arrays <em>can</em> cause havoc. I say can because you might inadvertently access memory you&#8217;re not supposed to, or you may access completely valid data, but not what you were expecting (depending on how the stack/heap is organized, compiler optimizations, kernel settings, and many other things).</p>
<p>We&#8217;ve all done this one. Using <code>&lt;=</code> versus <code>&lt;</code> can make all the difference. Blindly accepting user input as an array index can also lead to problems.</p>
<p>In some languages, like ruby, it might not be that big of a deal. If you index an array with an index value that is out of bounds, it just returns <code>nil</code>, so as long as you deal with that as a return value, you&#8217;re probably good.</p>
<p>Essentially, this error can lead to the standard <a href="http://en.wikipedia.org/wiki/Buffer_overflow">buffer overflow</a>, which can lead to an attacker executing their own code, and doing all sorts of nasty things. Try to keep your arrays in check.</p>
<h3>Ways around it</h3>
<p>Ways around this error are definitely language dependent. In Java, depending on what data structure you are working with, a <code>ArrayOutOfBoundsException</code>, <code>IndexOutOfBoundsException</code>, or other exception <em>may</em> be raised, which you could catch and deal with. Ruby simply returns <code>nil</code>, somebody you can also deal with gracefully. Lower level languages like C <em>may</em> continue to work fine, but may also explode and die in a (sometimes literal) fire.</p>
<p>In those scenarios, you&#8217;ll want to validate your input before indexing the array in the first place, and dealing with incorrect input appropriately.</p>
<p>This problem is really solved by the practice of sanitizing your inputs. Doing that will reduce your Tylenol bill.</p>
<h2><a href="http://cwe.mitre.org/data/definitions/98.html" title="&#39;PHP File Inclusion&#39;">13. Improper Control of Filename for Include/Require Statement in <span class="caps">PHP</span> Program</a></h2>
<p>As the title says, this is specific to <span class="caps">PHP</span>.</p>
<p>Okay so that&#8217;s not entirely true. While the actual weakness on the <span class="caps">CWE</span> site is specific to <span class="caps">PHP</span>, you can have the same problem with ruby, or really any other language that allows dynamic code loading.</p>
<p>If you are using user input to load files, and by that I mean using the input directly in a <code>require</code> or <code>include</code> statement, an attacker can pretty much do whatever they want.</p>
<h3>Ways around it</h3>
<p>Don&#8217;t do it? That seems like a pretty solid way around it.</p>
<p>Other ways are to at least validate the input. If you are expecting the value to be a theme name, check that the theme exists in the proper directory. This gets around directory traversal problems.</p>
<p>Specific to <span class="caps">PHP</span>, in your <code>php.ini</code> file, you can set <code>allow_url_fopen</code> to <code>false</code> to prevent remote files from being included. In ruby, remote files aren&#8217;t a problem since <code>require</code> and <code>load</code> only deal with files on disk.</p>
<h2><a href="http://cwe.mitre.org/data/definitions/805.html">12. Buffer Access with Incorrect Length Value</a></h2>
<p><figure><img src="http://cdn.verboselogging.com/transloadit/small/9c/c8b6fe7527d60caccfcaf5de36035b/spaghetti-os.jpg" class="fright bleft bbottom round" alt="" /></figure></p>
<p>Uh oh! Another buffer overflow problem. These are so common, and potentially so dangerous, but they don&#8217;t get the respect they deserve.</p>
<p>Anyway. This type of buffer overflow problem comes from using incorrect values and making assumptions. It&#8217;s always those stupid assumptions that get you! I&#8217;m going to use their example:</p>
<div class='clear'></div>
<script type='text/javascript' src='http://gist.github.com/418817.js?file=bad_length.c'></script><p>Only 64 bytes for a hostname? That&#8217;s not that much when it comes down to it, and if you do end up with a hostname longer than 63 characters (that last one is for the null terminator), the <code>strcpy</code> is going to end badly.</p>
<h3>Ways around it</h3>
<p>In this specific example, you should be using the safe variation, <code>strncpy</code>:</p>
<pre>strncpy(hostname, hp-&gt;h_name, 63); /* Leave 1 byte for null terminator */</pre>
<p>In this case, the <code>hostname</code> might not contain the correct (entire) hostname, but at least nothing explodes.</p>
<p>If you are using a language without such strict memory allocation requirements, you probably don&#8217;t have to worry about this kind of stuff.</p>
<p>I also like the <span class="caps">CWE</span> potential mitigations under the &#8216;Operation&#8217; heading: ensuring <a href="http://en.wikipedia.org/wiki/Data_execution_prevention">Data Execution Prevention</a> and <a href="http://en.wikipedia.org/wiki/ASLR">address space layout randomization</a> are enabled if available. As they point out, they aren&#8217;t complete catch all solutions, though they do make it much harder for attackers to do anything, so even if there is a buffer access problem, it will hopefully just crash the application, and not pose a huge security hole. <em>Hopefully</em>.</p>
<h2><a href="http://cwe.mitre.org/data/definitions/798.html">11. Use of Hard-coded Credentials</a></h2>
<p>I&#8217;ve done this. Granted, only for one-off scripts, but it&#8217;s not good. If you write code like this:</p>
<pre>MyDatabase.connect('localhost', 'theuser', 'thepassword');</pre>
<p>You&#8217;re doing it wrong. If you&#8217;re writing code like this:</p>
<pre>RemoteService.getData('remoteuser', 'remotepassword');</pre>
<p>You&#8217;re doing it wrong.</p>
<p>Connecting to your main application datastore, or another remote service with hardcoded credentials is just bad. You can dump the strings of a binary file, or decompile a Java class to get the strings in it. Ruby? Well it&#8217;s just plain text. Python? Plain text. Compiled python files (.pyc)? You could probably dump strings from those, and you can decompile them as well.</p>
<p>Basically, putting usernames and passwords in your source code is just bad. As soon as you do, they end up in version control, and they they are hanging out in your version control history forever. Not cool.</p>
<h3>Ways around it</h3>
<p>Don&#8217;t do it. Simple. Use a config file if you need passwords for anything.</p>
<p>If you are accepting passwords (like a web application), you&#8217;ll want to store a salted and hashed version of the passwords your users feed you. Storing passwords in cleartext in the database is for suckers.</p>
