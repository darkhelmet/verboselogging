--- 
id: 452
author: Daniel Huckstep
title: Super MongoDB MapReduce Max Out!
category: programming
description: I smack my data up with some MapReduce, courtesy of MongoDB.
published: true
publishedon: 22 Mar 2010 08:00 MDT
slugs: 
- super-mongodb-mapreduce-max-out
tags: 
- mongodb
- map-reduce
images: 
  mongodb: 
    original: http://cdn.verboselogging.com/transloadit/original/14/03bd5f35cb28de55a0aa0423e936a9/mongodb.png
    small: http://cdn.verboselogging.com/transloadit/small/02/619d28f5750684ac1e7ed8ce82b3bc/mongodb.png
    medium: http://cdn.verboselogging.com/transloadit/medium/c7/61f246fe2bd12b719c1aa056f8b4ed/mongodb.png
    large: http://cdn.verboselogging.com/transloadit/large/2b/e9476795b800d893cd6d4a16f3d6e2/mongodb.png
---
<p><figure><img src="http://cdn.verboselogging.com/transloadit/original/14/03bd5f35cb28de55a0aa0423e936a9/mongodb.png" class="fright bleft bbottom round original" alt="" /></figure></p>
<p>I&#8217;ve been playing with <a href="http://www.mongodb.org/">MongoDB</a> lately, and I must say, it&#8217;s the shit. In case you haven&#8217;t heard of MongoDB, let&#8217;s drop some buzz words:</p>
<ul>
	<li>Document oriented</li>
	<li>Dynamic queries</li>
	<li>Index support</li>
	<li>Replication support</li>
	<li>Query profiling</li>
	<li>MapReduce</li>
	<li>Auto sharding</li>
</ul>
<p>There are some more things, so check out their website for the full meal deal. I&#8217;m going to talk about the MapReduce part of things.</p>
<h2><a href="http://en.wikipedia.org/wiki/MapReduce">MapReduce</a></h2>
<p>The idea behind MapReduce has been around for a while; since the Lisp days. Here&#8217;s the basic idea:</p>
<ol>
	<li>Gather list of items (list 1).</li>
	<li>Apply the map function to each item in list 1, generating a new list (list 2).</li>
	<li>Apply the reduce function to the resultant list (list 2) as a whole.</li>
	<li>Return value return by reduce.</li>
	<li>Profit!</li>
</ol>
<p>In the MongoDB world, you run the mapReduce command, and it takes a few arguments:</p>
<ul>
	<li><strong>mapFunction</strong>
	<ul>
		<li>A function that takes an individual document (<code>{ "value": 1 }</code>) and (possibly) emits a value (or emit multiple values), whether that be a new document, or a single value (like a number).</li>
		<li>The emit function takes a key, and a value.</li>
	</ul></li>
</ul>
<script type="text/javascript" src="http://gist.github.com/339677.js?file=mongo-map.js"></script><ul>
	<li><strong>reduceFunction</strong>
	<ul>
		<li>A function that takes a list of values emitted from the map function and a key, and produces a single value.</li>
	</ul></li>
</ul>
<script type="text/javascript" src="http://gist.github.com/339677.js?file=mongo-reduce.js"></script><ul>
	<li><strong>optional options</strong>
	<ul>
		<li><strong>query</strong>
		<ul>
			<li>A MongoDB style query. Like any database query, this selects which documents you are going to apply your map function to.</li>
		</ul></li>
		<li><strong>out collection</strong>
		<ul>
			<li>The name of a collection to output into.</li>
		</ul></li>
		<li><strong>finalize function</strong>
		<ul>
			<li>A function to further apply to the reduced value.</li>
		</ul></li>
	</ul></li>
</ul>
<p>Here&#8217;s an example from the mongo shell.</p>
<script type="text/javascript" src="http://gist.github.com/339677.js?file=mongo-example.js"></script><p>So at the bottom there, you can see the result is 60.</p>
<p>We could rewrite this to move the <code>if</code> statement in the map function into a query. Then we cover less items, and don&#8217;t have to do the check in the map function.</p>
<script type="text/javascript" src="http://gist.github.com/339677.js?file=mongo-example-query.js"></script><p>It returns the same result as above.</p>
<p>With me so far? MapReduce is interesting if you&#8217;ve never seen it before or never done any functional programming, but once you get it, you understand its power.</p>
<h2>Caveats</h2>
<p>In the MongoDB environment, it&#8217;s incredibly important that your reduce function is idempotent. Stealing their example straight from the MongoDB website, it means:</p>
<pre>for all k,vals : reduce( k, [reduce(k,vals)] ) == reduce(k,vals)</pre>
<p>This is because the reduce function might be executed a number of times with results from various stages. Since MapReduce can be done across multiple servers, they will run their map and subsequent reduce functions on their data, but then the master server has to further reduce those results, so it takes the return values from all the reduce functions, and puts them into a list, and passes that to the reduce function again.</p>
<p>Basically, make sure the structure of what you return from reduce, is the same structure as whatever you are emitting in the map function. If you emit an integer, reduce should return an integer as well. In the sum example, it&#8217;s really straight forward in that we just add stuff up. In other situations it can get more complicated.</p>
<p>Next, I&#8217;ll talk about getting MapReduce to do, you know, useful things. Stay tuned!</p>
