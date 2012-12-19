--- 
id: 454
author: Daniel Huckstep
title: Writing User Defined Functions For Pig
category: programming
description: Pig is a powerful language for processing data. I show you how to leverage Java to write custom UDFs to help you out.
published: true
publishedon: 31 Mar 2010 08:00 MDT
slugs: 
- writing-user-defined-functions-for-pig
tags: 
- pig
- java
- hadoop
images: 
  pig: 
    original: http://cdn.verboselogging.com/transloadit/original/3c/da79d4fd788979554ab06a1c11306b/pig.jpg
    large: http://cdn.verboselogging.com/transloadit/large/01/33d69c0b0a67bcc695f0ea94e5fce9/pig.jpg
    small: http://cdn.verboselogging.com/transloadit/small/f9/9801c101f570f638d27f2682c76664/pig.jpg
    medium: http://cdn.verboselogging.com/transloadit/medium/ed/1e386bc5d83835a6fe66bb4d9cc7e3/pig.jpg
---
<p><figure><img src="http://cdn.verboselogging.com/transloadit/medium/ed/1e386bc5d83835a6fe66bb4d9cc7e3/pig.jpg" class="fright bleft bbottom round medium" alt="" /></figure></p>
<p>If you are processing a bunch of data, grouping it, joining it, filtering it, then you should probably be using <a href="http://hadoop.apache.org/pig/">pig</a>.</p>
<p>So go download that, and get it all setup. You need:</p>
<ul>
	<li>Java 1.6 (with <code>JAVA_HOME</code> setup)</li>
	<li><a href="http://hadoop.apache.org/common/">Hadoop</a> (with <code>HADOOP_HOME</code> setup)</li>
	<li>pig (of course)</li>
</ul>
<p>Put all the relevant stuff in your <code>PATH</code> too.</p>
<div class='clear'></div>
<h2>pig 101</h2>
<p>So here&#8217;s a simple pig script.</p>
<script type="text/javascript" src="http://gist.github.com/348301.js?file=example.pig"></script><p>This registers a jar file and defines a custom <acronym title="User Defined Function"><span class="caps">UDF</span></acronym> for doing whatever. It happens to be a log line parser function.</p>
<p>We load a bzipped log file from apache (it can just read the bzipped files! Wee!) and by using the <code>TextLoader</code>, each line comes in as a <code>chararray</code> (in pig terms, a string).</p>
<p>Now, <code>FOREACH</code> line, run it through the parser function we defined ealier. We&#8217;ll look at this shortly. We can now do some fun stuff, like <code>GROUP</code> on the action, and generate the counts of all these things.</p>
<p>Okay so that might look a little weird, but if you read it, it makes perfect sense. Let&#8217;s cover a few things before we get to the <span class="caps">UDF</span> fun.</p>
<h2><code>FOREACH</code> and <code>GENERATE</code></h2>
<p>In pig, the <code>FOREACH</code> and <code>GENERATE</code> combination does sort of what it says. It&#8217;s essentially the map function (and you should be familiar with map functions from the <a href="/2010/03/22/super-mongodb-mapreduce-max-out">previous</a> <a href="/2010/03/28/finally-mapreduce-for-profit">posts</a>). For every <em>thing</em> in the bag (a bag is a pig datatype), <em>generate</em> something. In this case, we are telling pig to use our custom class to take the line, and generate some stuff (a tuple, actually).</p>
<h2>Tuples and Schemas</h2>
<p>Tuples are ordered groups of things, and in pig, the fields can be named. You see tuples in Haskell, lisp (I think), and other programming languages. In the scripts, the <code>logs</code> variable represents a bunch of tuples, where each tuple is a single item, and that single item is named <em>line</em>. We got this because when we said:</p>
<pre>logs = LOAD 'apache.log.bz2' USING TextLoader AS (line: chararray);</pre>
<p>It&#8217;s telling pig</p>
<blockquote>
<p>Load the file and treat it as a text file, splitting on newlines, and give me a bunch of tuples, where each tuple has a single item that is a chararray named line.</p>
</blockquote>
<p>You could load the file the same way, omitting the <code>AS (line: chararray)</code> part, but then the resulting tuples would have no <em>schema</em>.</p>
<p>The schema is essentially type information about the tuple. You can have a tuple without a schema, but it&#8217;s much more useful to have one, since you can refer to field by name, instead of by field number (like indexing an array).</p>
<h2>User Defined Functions</h2>
<p>A User Defined Function is exactly that; it&#8217;s something you write that pig loads and uses. In this case, we are writing a Java class (Java is the only language you can use for this currently). For this example, we are going to write a function to parse a line in a log file and return a tuple so pig can then work its magic with the tuples. So normally this takes 30 minutes to bake, but I&#8217;ve got one already in the oven!</p>
<script type="text/javascript" src="http://gist.github.com/348301.js?file=LogParser.java"></script><h2>Play by play</h2>
<p>Okay, first of all remember to add to your classpath the pig jar file. It&#8217;s the <code>pig-VERSION-core.jar</code> in the pig directory. Add it in Eclipse, or whatever, so when you compile it has access to everything.</p>
<p>Inherit your class from <code>EvalFunc&lt;Tuple&gt;</code> since that&#8217;s exactly what we are making: an <code>EvalFunc</code> (as opposed to a filter function or something else) that returns a tuple.</p>
<p>The exec method is your main method that has to return the proper type (tuple in our case) and takes a tuple. We check to ensure the input tuple is nice, in that it exists and has only one item (the line of text). We can then get the first item and cast it to a <code>String</code> so we can work with it.</p>
<p>We use a <code>try/catch</code> block to handle errors and make sure we just return <code>null</code> if there are any problems. If you return null, <a href="http://stackoverflow.com/questions/2540071/does-throwing-an-exception-in-an-evalfunc-pig-udf-skip-just-that-line-or-stop-co/2541842#2541842">everything in the tuple is null</a> so you can filter that out using standard pig stuff.</p>
<p>We use the <code>TupleFactory</code> singleton to get a tuple, append our values in the order we want them to appear (in this case, just the <span class="caps">HTTP</span> method, IP address, and date), and return it. Yay!</p>
<h2>I can haz schema?</h2>
<p>Yes you can. You could write the schema in the pig script.</p>
<pre>log_events = FOREACH logs GENERATE FLATTEN(Parser(line)) AS (action: chararray, ip: chararray, date: chararray);</pre>
<p>This does have benefits, the main one being the schema is right there and you can see it. This makes writing the rest of the script a little easier, since you don&#8217;t have to remember exactly what&#8217;s in the tuples. The downside is you have to change code in two separate spots.</p>
<p>We decided to put the schema in the java class, so you can do some more programmatic things with it, and when you have to change it, it&#8217;s right there next to the <code>exec</code> method you are also changing.</p>
<p>Building a schema is a little epic in Java (verbose much?) but it&#8217;s not terrible. We create a new schema, and add in the same order we added things in the <code>exec</code> method, the names and types of the things we added.</p>
<ol>
	<li><span class="caps">HTTP</span> method: String/chararray</li>
	<li>IP address: String/chararray</li>
	<li>Date: String/chararray</li>
</ol>
<p>You add to the new schema a <code>Schema.FieldSchema</code> object where you specify the name (what you want to reference the field as in pig) and the type (a byte, but just use the <code>DataType</code> enum values).</p>
<p>Now, if you <code>DESCRIBE log_events;</code> in the pig shell, it will tell you the schema. You can also now use named indexes into the tuple, as with <code>GROUP log_events BY action</code> to make your code more readable.</p>