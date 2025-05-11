[](https://github.com/arXiv/arxiv-docs/blob/develop/source/help/api/user-manual.md "Edit this page")

![API logo](https://info.arxiv.org/help/api/arXiv_api_xml.png)

Please review the [Terms of Use for arXiv APIs](https://info.arxiv.org/help/api/tou.html) before using the arXiv API.

### Table of Contents

[1\. Preface](https://info.arxiv.org/help/api/user-manual.html#_preface)  
[2\. API QuickStart](https://info.arxiv.org/help/api/user-manual.html#Quickstart)  
[3\. Structure of the API](https://info.arxiv.org/help/api/user-manual.html#Architecture)  
[3.1. Calling the API](https://info.arxiv.org/help/api/user-manual.html#_calling_the_api)  
[3.1.1. Query Interface](https://info.arxiv.org/help/api/user-manual.html#_query_interface)  
[3.1.1.1. search\_query and id\_list logic](https://info.arxiv.org/help/api/user-manual.html#search_query_and_id_list)  
[3.1.1.2. start and max\_results paging](https://info.arxiv.org/help/api/user-manual.html#paging)  
[3.1.1.3. sort order for return results](https://info.arxiv.org/help/api/user-manual.html#sort)  
[3.2. The API Response](https://info.arxiv.org/help/api/user-manual.html#api_response)  
[3.3. Outline of an Atom feed](https://info.arxiv.org/help/api/user-manual.html#atom_feed_outline)  
[3.3.1. Feed Metadata](https://info.arxiv.org/help/api/user-manual.html#_feed_metadata)  
[3.3.1.1. <title>, <id>, <link> and <updated>](https://info.arxiv.org/help/api/user-manual.html#_lt_title_gt_lt_id_gt_lt_link_gt_and_lt_updated_gt)  
[3.3.1.2. OpenSearch Extension Elements](https://info.arxiv.org/help/api/user-manual.html#_opensearch_extension_elements)  
[3.3.2. Entry Metadata](https://info.arxiv.org/help/api/user-manual.html#_entry_metadata)  
[3.3.2.1. <title>, <id>, <published>, and <updated>](https://info.arxiv.org/help/api/user-manual.html#title_id_published_updated)  
[3.3.2.1. <summary>, <author> and <category>](https://info.arxiv.org/help/api/user-manual.html#_lt_summary_gt_lt_author_gt_and_lt_category_gt)  
[3.3.2.3. <link>'s](https://info.arxiv.org/help/api/user-manual.html#entry_links)  
[3.3.2.4. <arxiv> extension elements](https://info.arxiv.org/help/api/user-manual.html#extension_elements)  
[3.4. Errors](https://info.arxiv.org/help/api/user-manual.html#errors)  
[4\. Examples](https://info.arxiv.org/help/api/user-manual.html#Examples)  
[4.1. Simple Examples](https://info.arxiv.org/help/api/user-manual.html#_simple_examples)  
[4.1.1. Perl](https://info.arxiv.org/help/api/user-manual.html#perl_simple_example)  
[4.1.2. Python](https://info.arxiv.org/help/api/user-manual.html#python_simple_example)  
[4.1.3. Ruby](https://info.arxiv.org/help/api/user-manual.html#ruby_simple_example)  
[4.1.4. PHP](https://info.arxiv.org/help/api/user-manual.html#php_simple_example)  
[4.2. Detailed Parsing Examples](https://info.arxiv.org/help/api/user-manual.html#detailed_examples)  
[5\. Appendices](https://info.arxiv.org/help/api/user-manual.html#Appendices)  
[5.1. Details of Query Construction](https://info.arxiv.org/help/api/user-manual.html#query_details)  
[5.1.1. A Note on Article Versions](https://info.arxiv.org/help/api/user-manual.html#_a_note_on_article_versions)  
[5.2. Details of Atom Results Returned](https://info.arxiv.org/help/api/user-manual.html#_details_of_atom_results_returned)  
[5.3. Subject Classifications](https://info.arxiv.org/help/api/user-manual.html#subject_classifications)

## 1\. Preface

The arXiv API allows programmatic access to the hundreds of thousands of e-prints hosted on [arXiv.org](http://arxiv.org/).

This manual is meant to provide an introduction to using the API, as well as documentation describing its details, and as such is meant to be read by both beginning and advanced users. To get a flavor for how the API works, see the [API Quickstart](https://info.arxiv.org/help/api/user-manual.html#Quickstart). For more detailed information, see [Structure of the API](https://info.arxiv.org/help/api/user-manual.html#Architecture).

For examples of using the API from several popular programming languages including perl, python and ruby, see the [Examples](https://info.arxiv.org/help/api/user-manual.html#Examples) section.

Finally, the [Appendices](https://info.arxiv.org/help/api/user-manual.html#Appendices) contain an explanation of all input parameters to the API, as well as the output format.

## 2\. API QuickStart

The easiest place to start with the API is by accessing it through a web browser. For examples of accessing the API through common programming languages, see the [Examples](https://info.arxiv.org/help/api/user-manual.html#Examples) section.

Most everyone that has read or submitted e-prints on the arXiv is familiar with the arXiv human web interface. These HTML pages can be accessed by opening up your web browser, and entering the following url in your web browser

[http://arxiv.org](http://arxiv.org/)

From there, the article listings can be browsed by clicking on one of the many links, or you can search for articles using the search box in the upper right hand side of the page. For example, if I wanted to search for articles that contain the word `electron` in the title or abstract, I would type `electron` in the search box, and click `Go`. If you follow my example, you will see [something like this](http://arxiv.org/find/all/1/all:+electron/0/1/0/all/0/1): a web page listing the title and authors of each result, with links to the abstract page, pdf, etc.

In its simplest form, the API can be used in exactly the same way. However, it uses a few shortcuts so there is less clicking involved. For example, you can see the same search results for `electron` by entering the url

[http://export.arxiv.org/api/query?search\_query=all:electron](http://export.arxiv.org/api/query?search_query=all:electron).

Alternatively, you can search for articles that contain `electron` _AND_ `proton` with the API by entering

[http://export.arxiv.org/api/query?search\_query=all:electron+AND+all:proton](http://export.arxiv.org/api/query?search_query=all:electron+AND+all:proton)

What you see will look different from the HTML interface, but it contains the same information as the search done with the human interface. The reason why the results look different is that the API returns results in the Atom 1.0 format, and not HTML. Since Atom is defined as an XML grammar, it is much easier to digest for programs than HTML. The API is not intended to be used inside a web browser by itself, but this is a particularly simple way to debug a program that does use the API.

You might notice that your web browser has asked you if you want to “subscribe to this feed” after you enter the API url. This is because Atom is one of the formats used by web sites to syndicate their content. These feeds are usually read with feed reader software, and are what is generated by the existing [arXiv rss feeds](http://arxiv.org/help/rss). The current arXiv feeds only give you updates on new papers within the category you specify. One immediately useful thing to do with the API then is to generate your own feed, based on a custom query!

To learn more about how to construct custom search queries with the API, see the appendix on the [details of query construction](https://info.arxiv.org/help/api/user-manual.html#query_details). To learn about what information is returned by the API, see the section on [the API response](https://info.arxiv.org/help/api/user-manual.html#api_response). To learn more about writing programs to call the API, and digest the responses, we suggest starting with the section on [Structure of the API](https://info.arxiv.org/help/api/user-manual.html#Architecture).

## 3\. Structure of the API

In this section, we'll go over some of the details of interacting with the API. A diagram of a typical API call is shown below:

**Example: A typical API call**

```
Request from url: http://export.arxiv.org/api/query  (1)
 with parameters: search_query=all:electron
                .
                .
                .
API server processes the request and sends the response
                .
                .
                .
Response received by client.  (2)
```

1.  The request can be made via HTTP GET, in which the parameters are encoded in the url, or via an HTTP POST in which the parameters are encoded in the HTTP request header. Most client libraries support both methods.
    
2.  If all goes well, the HTTP header will show a 200 OK status, and the response body will contain the Atom response content as shown in the [example response](https://info.arxiv.org/help/api/user-manual.html#response_example).
    

### 3.1. Calling the API

As mentioned above, the API can be called with an HTTP request of type GET or POST. For our purposes, the main difference is that the parameters are included in the url for a GET request, but not for the POST request. Thus if the parameters list is unusually long, a POST request might be preferred.

The parameters for each of the API methods are explained below. For each method, the base url is

```
http://export.arxiv.org/api/{method_name}?{parameters}
```

#### 3.1.1. Query Interface

The API query interface has `method_name=query`. The table below outlines the parameters that can be passed to the query interface. Parameters are separated with the `&` sign in the constructed url's.

##### 3.1.1.1. search\_query and id\_list logic

We have already seen the use of `search_query` in the [quickstart](https://info.arxiv.org/help/api/user-manual.html#Quickstart) section. The `search_query` takes a string that represents a search query used to find articles. The construction of `search_query` is described in the [search query construction appendix](https://info.arxiv.org/help/api/user-manual.html#query_details). The `id_list` contains a comma-delimited list of arXiv id's.

The logic of these two parameters is as follows:

-   If only `search_query` is given (`id_list` is blank or not given), then the API will return results for each article that matches the search query.
    
-   If only `id_list` is given (`search_query` is blank or not given), then the API will return results for each article in `id_list`.
    
-   If _BOTH_ `search_query` and `id_list` are given, then the API will return each article in `id_list` that matches `search_query`. This allows the API to act as a results filter.
    

This is summarized in the following table:

##### 3.1.1.2. start and max\_results paging

Many times there are hundreds of results for an API query. Rather than download information about all the results at once, the API offers a paging mechanism through `start` and `max_results` that allows you to download chucks of the result set at a time. Within the total results set, `start` defines the index of the first returned result, _using 0-based indexing_. `max_results` is the number of results returned by the query. For example, if wanted to step through the results of a `search_query` of `all:electron`, we would construct the urls:

```
http://export.arxiv.org/api/query?search_query=all:electron&amp;start=0&amp;max_results=10 (1)
http://export.arxiv.org/api/query?search_query=all:electron&amp;start=10&amp;max_results=10 (2)
http://export.arxiv.org/api/query?search_query=all:electron&amp;start=20&amp;max_results=10 (3)
```

1.  Get results 0-9
    
2.  Get results 10-19
    
3.  Get results 20-29
    

Detailed examples of how to perform paging in a variety of programming languages can be found in the [examples](https://info.arxiv.org/help/api/user-manual.html#detailed_examples) section.

In cases where the API needs to be called multiple times in a row, we encourage you to play nice and incorporate a 3 second delay in your code. The [detailed examples](https://info.arxiv.org/help/api/user-manual.html#detailed_examples) below illustrate how to do this in a variety of languages.

Because of speed limitations in our implementation of the API, the maximum number of results returned from a single call (`max_results`) is limited to 30000 in slices of at most 2000 at a time, using the `max_results` and `start` query parameters. For example to retrieve matches 6001-8000: http://export.arxiv.org/api/query?search\_query=all:electron&start=6000&max\_results=2000

Large result sets put considerable load on the server and also take a long time to render. We recommend to refine queries which return more than 1,000 results, or at least request smaller slices. For bulk metadata harvesting or set information, etc., the [OAI-PMH](https://info.arxiv.org/help/oa/index.html) interface is more suitable. A request with `max_results` >30,000 will result in an HTTP 400 error code with appropriate explanation. A request for 30000 results will typically take a little over 2 minutes to return a response of over 15MB. Requests for fewer results are much faster and correspondingly smaller.

##### 3.1.1.3. sort order for return results

There are two options for for the result set to the API search, `sortBy` and `sortOrder`.

`sortBy` can be "relevance" (Apache Lucene's default [RELEVANCE](https://lucene.apache.org/core/3_0_3/api/core/org/apache/lucene/search/Sort.html) ordering), "lastUpdatedDate", "submittedDate"

`sortOrder` can be either "ascending" or "descending"

A sample query using these new parameters looks like:

```
http://export.arxiv.org/api/query?search_query=ti:"electron thermal conductivity"&amp;sortBy=lastUpdatedDate&amp;sortOrder=ascending
```

### 3.2. The API Response

Everything returned by the API in the body of the HTTP responses is Atom 1.0, including [errors](https://info.arxiv.org/help/api/user-manual.html#errors). Atom is a grammar of XML that is popular in the world of content syndication, and is very similar to RSS for this purpose. Typically web sites with dynamic content such as news sites and blogs will publish their content as Atom or RSS feeds. However, Atom is a general format that embodies the concept of a list of items, and thus is well-suited to returning the arXiv search results.

### 3.3. Outline of an Atom feed

In this section we will discuss the contents of the Atom documents returned by the API. To see the full explanation of the Atom 1.0 format, please see the [Atom specification](http://www.ietf.org/rfc/rfc4287.txt).

An API response consists of an Atom `<feed>` element which contains metadata about the API call performed, as well as child `<entry>` elements which embody the metadata for each of the returned results. Below we explain each of the elements and attributes. We will base our discussion on the [sample results feed](https://info.arxiv.org/help/api/user-manual.html#response_example) discussed in the examples section.

<sup>You may notice that the results from the API are ordered differently that the results given by the <a href="http://arxiv.org/find">HTML arXiv search interface</a>. The HTML interface automatically sorts results in descending order based on the date of their submission, while the API returns results according to relevancy from the internal search engine. Thus when debugging a search query, we encourage you to use the API within a web browser, rather than the HTML search interface. If you want sorting by date, you can always do this within your programs by reading the <code>&lt;published&gt;</code> tag for each entry as explained <a href="https://info.arxiv.org/help/api/user-manual.html#title_id_published_updated">below</a>.</sup>

#### 3.3.1. Feed Metadata

Every response will contain the line:

```
&lt;?xml version="1.0" encoding="utf-8"?&gt;
```

to signify that we are receiving XML 1.0 with a UTF-8 encoding. Following that line will be a line indicating that we are receiving an Atom feed:

```
&lt;feed xmlns="http://www.w3.org/2005/Atom"
xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"
xmlns:arxiv="http://arxiv.org/schemas/atom"&gt;
```

You will notice that three XML namespaces are defined. The default namespace signifies that we are dealing with Atom 1.0. The other two namespaces define extensions to Atom that we describe below.

##### 3.3.1.1. <title>, <id>, <link> and <updated>

The `<title>` element gives the title for the feed:

```
&lt;title xmlns="http://www.w3.org/2005/Atom"&gt;
    ArXiv Query:  search_query=all:electron&amp;amp;id_list=&amp;amp;start=0&amp;amp;max_results=1
&lt;/title&gt;
```

The title contains a canonicalized version of the query used to call the API. The canonicalization includes all parameters, using their defaults if they were not included, and always puts them in the order `search_query`,`id_list`,`start`,`max_results`, even if they were specified in a different order in the actual query.

The `<id>` element serves as a unique id for this query, and is useful if you are writing a program such as a feed reader that wants to keep track of all the feeds requested in the past. This id can then be used as a key in a database.

```
&lt;id xmlns="http://www.w3.org/2005/Atom"&gt;
    http://arxiv.org/api/cHxbiOdZaP56ODnBPIenZhzg5f8
&lt;/id&gt;
```

The id is guaranteed to be unique for each query.

The `<link>` element provides a URL that can be used to retrieve this feed again.

```
&lt;link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/api/query?search_query=all:electron&amp;amp;id_list=&amp;amp;start=0&amp;amp;max_results=1" rel="self" type="application/atom+xml"/&gt;
```

Note that the url in the link represents the canonicalized version of the query. The `<link>` provides a GET requestable url, even if the original request was done via POST.

The `<updated>` element provides the last time the contents of the feed were last updated:

```
&lt;updated xmlns="http://www.w3.org/2005/Atom"&gt;2007-10-08T00:00:00-04:00&lt;/updated&gt;
```

<sup>Because the arXiv submission process works on a 24 hour submission cycle, new articles are only available to the API on the midnight <em>after</em> the articles were processed. The <code>&lt;updated&gt;</code> tag thus reflects the midnight of the day that you are calling the API. <strong>This is very important</strong> - search results do not change until new articles are added. Therefore there is no need to call the API more than once in a day for the same query. Please cache your results. This primarily applies to production systems, and of course you are free to play around with the API while you are developing your program!</sup>

##### 3.3.1.2. OpenSearch Extension Elements

There are several extension elements defined in the OpenSearch namespace

```
http://a9.com/-/spec/opensearch/1.1/
```

[OpenSearch](http://www.opensearch.org/Home) is a lightweight technology that acts in a similar way as the Web Services Description Language. The OpenSearch elements we have included allow OpenSearch enabled clients to digest our results. Such clients often include search result aggregators and browser pluggins that allow searching from a variety of sources.

The OpenSearch extension elements can still be useful to you even if you are not writing one of these applications. The `<opensearch:totalResults>` element lists how many results are in the result set for the query:

```
&lt;opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;
   1000
&lt;/opensearch:totalResults&gt;
```

This can be very useful when implementing [paging of search results](https://info.arxiv.org/help/api/user-manual.html#paging). The other two elements `<opensearch:startIndex>`, and `<opensearch:itemsPerPage>` are analogous to `start`, and `max_results` [discussed above](https://info.arxiv.org/help/api/user-manual.html#paging).

```
&lt;opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;
   0
&lt;/opensearch:startIndex&gt;
&lt;opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;
   1
&lt;/opensearch:itemsPerPage&gt;
```

#### 3.3.2. Entry Metadata

If there are no errors, the `<feed>` element contains 0 or more child `<entry>` elements with each `<entry>` representing an article in the returned results set. As explained in the [errors](https://info.arxiv.org/help/api/user-manual.html#errors) section, if there are errors, a single `<entry>` element representing the error is returned. Below the element description describes the elements for `<entry>`'s representing arXiv articles. For a general discussion of arXiv metadata, see the [arXiv metadata explanation](https://info.arxiv.org/help/prep.html).

##### 3.3.2.1. <title>, <id>, <published>, and <updated>

The `<title>` element contains the title of the article returned:

```
&lt;title xmlns="http://www.w3.org/2005/Atom"&gt;
    Multi-Electron Production at High Transverse Momenta in ep Collisions at HERA
&lt;/title&gt;
```

The `<id>` element contains a url that resolves to the abstract page for that article:

```
&lt;id xmlns="http://www.w3.org/2005/Atom"&gt;
    http://arxiv.org/abs/hep-ex/0307015
&lt;/id&gt;
```

If you want only the arXiv id for the article, you can remove the leading `http://arxiv.org/abs/` in the `<id>`.

The `<published>` tag contains the date in which the `first` version of this article was submitted and processed. The `<updated>` element contains the date on which the retrieved article was submitted and processed. If the version is version 1, then `<published> == <updated>`, otherwise they are different. In the example below, the article retrieved was version 2, so `<updated>` and `<published>` are different (see the [original query](http://export.arxiv.org/api/query?id_list=cond-mat/0702661v2)).

```
&lt;published xmlns="http://www.w3.org/2005/Atom"&gt;
    2007-02-27T16:02:02-05:00
&lt;/published&gt;
&lt;updated xmlns="http://www.w3.org/2005/Atom"&gt;
    2007-06-25T17:09:59-04:00
&lt;/updated&gt;
```

The `<summary>` element contains the abstract for the article:

```
&lt;summary xmlns="http://www.w3.org/2005/Atom"&gt;
    Multi-electron production is studied at high electron transverse momentum
    in positron- and electron-proton collisions using the H1 detector at HERA.
    The data correspond to an integrated luminosity of 115 pb-1. Di-electron
    and tri-electron event yields are measured. Cross sections are derived in
    a restricted phase space region dominated by photon-photon collisions. In
    general good agreement is found with the Standard Model predictions.
    However, for electron pair invariant masses above 100 GeV, three
    di-electron events and three tri-electron events are observed, compared to
    Standard Model expectations of 0.30 \pm 0.04 and 0.23 \pm 0.04,
    respectively.
&lt;/summary&gt;
```

There is one `<author>` element for each author of the paper in order of authorship. Each `<author>` element has a `<name>` sub-element which contains the name of the author.

```
&lt;author xmlns="http://www.w3.org/2005/Atom"&gt;
      &lt;name xmlns="http://www.w3.org/2005/Atom"&gt;H1 Collaboration&lt;/name&gt;
&lt;/author&gt;
```

If author affiliation is present, it is included as an `<arxiv:affiliation>` subelement of the `<author>` element as discussed [below](https://info.arxiv.org/help/api/user-manual.html#extension_elements).

The `<category>` element is used to describe either an arXiv, ACM, or MSC classification. See the [arXiv metadata explanation](http://arxiv.org/help/prep) for more details about these classifications. The `<category>` element has two attributes, `scheme`, which is the categorization scheme, and `term` which is the term used in the categorization. Here is an example from the query [http://export.arxiv.org/api/query?id\_list=cs/9901002v1](http://export.arxiv.org/api/query?id_list=cs/9901002v1)

```
&lt;category xmlns="http://www.w3.org/2005/Atom" term="cs.LG" scheme="http://arxiv.org/schemas/atom"/&gt;
&lt;category xmlns="http://www.w3.org/2005/Atom" term="cs.AI" scheme="http://arxiv.org/schemas/atom"/&gt;
&lt;category xmlns="http://www.w3.org/2005/Atom" term="I.2.6" scheme="http://arxiv.org/schemas/atom"/&gt;
```

Note that in this example, there are 3 category elements, one for each category. The first two correspond to arXiv categories, and the last one to an ACM category. See [<arxiv> extension elements](https://info.arxiv.org/help/api/user-manual.html#extension_elements) below for information on how to identify the arXiv primary category.

##### 3.3.2.3. <link>'s

For each entry, there are up to three `<link>` elements, distinguished by their `rel` and `title` attributes. The table below summarizes what these links refer to

For example:

```
&lt;link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/abs/hep-ex/0307015v1" rel="alternate" type="text/html"/&gt;
&lt;link xmlns="http://www.w3.org/2005/Atom" title="pdf" href="http://arxiv.org/pdf/hep-ex/0307015v1" rel="related" type="application/pdf"/&gt;
&lt;link xmlns="http://www.w3.org/2005/Atom" title="doi" href="http://dx.doi.org/10.1529/biophysj.104.047340" rel="related"/&gt;
```

##### 3.3.2.4. <arxiv> extension elements

There are several pieces of [arXiv metadata](http://arxiv.org/help/prep) that are not able to be mapped onto the standard Atom specification. We have therefore defined several extension elements which live in the `arxiv` namespace

```
http://arxiv.org/schemas/atom
```

The arXiv classification system supports multiple <category> tags, as well as a primary classification. The primary classification is a replica of an Atom <category> tag, except it has the name `<arxiv:primary_category>`. For example, from the query [http://export.arxiv.org/api/query?id\_list=cs/9901002v1](http://export.arxiv.org/api/query?id_list=cs/9901002v1), we have

```
&lt;arxiv:primary_category xmlns:arxiv="http://arxiv.org/schemas/atom" term="cs.LG" scheme="http://arxiv.org/schemas/atom"/&gt;
```

signifying that `cs.LG` is the primary arXiv classification for this e-print.

The `<arxiv:comment>` element contains the typical author comments found on most arXiv articles:

```
&lt;arxiv:comment xmlns:arxiv="http://arxiv.org/schemas/atom"&gt;
   23 pages, 8 figures and 4 tables
&lt;/arxiv:comment&gt;
```

If the author has supplied affiliation information, then this is included as an `<arxiv:affiliation>` subelement of the standard Atom `<author>` element. For example, from the query [http://export.arxiv.org/api/query?id\_list=0710.5765v1](http://export.arxiv.org/api/query?id_list=0710.5765v1), we have

```
&lt;author&gt;
   &lt;name&gt;G. G. Kacprzak&lt;/name&gt;
   &lt;arxiv:affiliation xmlns:arxiv="http://arxiv.org/schemas/atom"&gt;NMSU&lt;/arxiv:affiliation&gt;
&lt;/author&gt;
```

If the author has provided a journal reference for the article, then there will be a `<arxiv:journal_ref>` element with this information:

```
&lt;arxiv:journal_ref xmlns:arxiv="http://arxiv.org/schemas/atom"&gt;
   Eur.Phys.J. C31 (2003) 17-29
&lt;/arxiv:journal_ref&gt;
```

If the author has provided a DOI for the article, then there will be a `<arxiv:doi>` element with this information:

```
&lt;arxiv:doi xmlns:arxiv="http://arxiv.org/schemas/atom"&gt;
   10.1529/biophysj.104.047340
&lt;/arxiv:doi&gt;
```

### 3.4. Errors

Errors are returned as Atom feeds with a single entry representing the error. The `<summary>` for the error contains a helpful error message, and the `<link>` element contains a url to a more detailed explanation of the message.

For example, the API call [http://export.arxiv.org/api/query?id\_list=1234.12345](http://export.arxiv.org/api/query?id_list=1234.12345) contains a malformed id, and results in the error

```
&lt;?xml version="1.0" encoding="utf-8"?&gt;
&lt;feed xmlns="http://www.w3.org/2005/Atom" xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;
  &lt;link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/api/query?search_query=&amp;amp;id_list=1234.12345" rel="self" type="application/atom+xml"/&gt;
  &lt;title xmlns="http://www.w3.org/2005/Atom"&gt;ArXiv Query: search_query=&amp;amp;id_list=1234.12345&lt;/title&gt;
  &lt;id xmlns="http://www.w3.org/2005/Atom"&gt;http://arxiv.org/api/kvuntZ8c9a4Eq5CF7KY03nMug+Q&lt;/id&gt;
  &lt;updated xmlns="http://www.w3.org/2005/Atom"&gt;2007-10-12T00:00:00-04:00&lt;/updated&gt;
  &lt;opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;1&lt;/opensearch:totalResults&gt;
  &lt;opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;0&lt;/opensearch:startIndex&gt;

  &lt;opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"&gt;1&lt;/opensearch:itemsPerPage&gt;
  &lt;entry xmlns="http://www.w3.org/2005/Atom"&gt;
    &lt;id xmlns="http://www.w3.org/2005/Atom"&gt;http://arxiv.org/api/errors#incorrect_id_format_for_1234.12345&lt;/id&gt;
    &lt;title xmlns="http://www.w3.org/2005/Atom"&gt;Error&lt;/title&gt;
    &lt;summary xmlns="http://www.w3.org/2005/Atom"&gt;incorrect id format for 1234.12345&lt;/summary&gt;
    &lt;updated xmlns="http://www.w3.org/2005/Atom"&gt;2007-10-12T00:00:00-04:00&lt;/updated&gt;

    &lt;link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/api/errors#incorrect_id_format_for_1234.12345" rel="alternate" type="text/html"/&gt;
    &lt;author xmlns="http://www.w3.org/2005/Atom"&gt;
      &lt;name xmlns="http://www.w3.org/2005/Atom"&gt;arXiv api core&lt;/name&gt;
    &lt;/author&gt;
  &lt;/entry&gt;
&lt;/feed&gt;
```

The following table gives information on errors that might occur.

## 4\. Examples

Once you have familiarized yourself with the API, you should be able to easily write programs that call the API automatically. Most programming languages, if not all, have libraries that allow you to make HTTP requests. Since Atom is growing, not all languages have libraries that support Atom parsing, so most of the programming effort will be in digesting the responses you receive. The languages that we know of that can easily handle calling the api via HTTP and parsing the results include:

-   [Perl](http://www.perl.org/) (via [LWP](http://search.cpan.org/~gaas/libwww-perl-5.808/lib/LWP.pm)) ([example](https://info.arxiv.org/help/api/user-manual.html#perl_simple_example))
    
-   [Python](http://www.python.org/) (via [urllib](https://docs.python.org/3/library/index.html)) ([example](https://info.arxiv.org/help/api/user-manual.html#python_simple_example))
    
-   [Ruby](http://www.ruby-lang.org/) (via [uri](https://ruby-doc.org/stdlib-2.5.1/libdoc/uri/rdoc/URI.html) and [net::http](https://ruby-doc.org/stdlib-2.7.0/libdoc/net/http/rdoc/Net/HTTP.html)) ([example](https://info.arxiv.org/help/api/user-manual.html#ruby_simple_example))
    
-   [PHP](http://www.php.net/) (via file\_get\_contents()) ([example](https://info.arxiv.org/help/api/user-manual.html#php_simple_example))
    

### 4.1. Simple Examples

Below we include code snippets for these languages that perform the bare minimum functionality - calling the api and printing the raw Atom results. If your favorite language is not up here, write us with an example, and we'll be glad to post it!

All of the simple examples produce an output which looks like:

Example: A Typical Atom Response

```atom
<?xml version="1.0" encoding="utf-8"?> <feed xmlns="http://www.w3.org/2005/Atom" xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/" xmlns:arxiv="http://arxiv.org/schemas/atom"> <link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/api/query?search_query=all:electron&amp;id_list=&amp;start=0&amp;max_results=1" rel="self" type="application/atom+xml"/> <title xmlns="http://www.w3.org/2005/Atom">ArXiv Query: search_query=all:electron&amp;id_list=&amp;start=0&amp;max_results=1</title> <id xmlns="http://www.w3.org/2005/Atom">http://arxiv.org/api/cHxbiOdZaP56ODnBPIenZhzg5f8</id> <updated xmlns="http://www.w3.org/2005/Atom">2007-10-08T00:00:00-04:00</updated> <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1000</opensearch:totalResults> <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex> <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:itemsPerPage> <entry xmlns="http://www.w3.org/2005/Atom" xmlns:arxiv="http://arxiv.org/schemas/atom"> <id xmlns="http://www.w3.org/2005/Atom">http://arxiv.org/abs/hep-ex/0307015</id> <published xmlns="http://www.w3.org/2005/Atom">2003-07-07T13:46:39-04:00</published> <updated xmlns="http://www.w3.org/2005/Atom">2003-07-07T13:46:39-04:00</updated> <title xmlns="http://www.w3.org/2005/Atom">Multi-Electron Production at High Transverse Momenta in ep Collisions at HERA</title> <summary xmlns="http://www.w3.org/2005/Atom"> Multi-electron production is studied at high electron transverse momentum in positron- and electron-proton collisions using the H1 detector at HERA. The data correspond to an integrated luminosity of 115 pb-1. Di-electron and tri-electron event yields are measured. Cross sections are derived in a restricted phase space region dominated by photon-photon collisions. In general good agreement is found with the Standard Model predictions. However, for electron pair invariant masses above 100 GeV, three di-electron events and three tri-electron events are observed, compared to Standard Model expectations of 0.30 \pm 0.04 and 0.23 \pm 0.04, respectively. </summary> <author xmlns="http://www.w3.org/2005/Atom"> <name xmlns="http://www.w3.org/2005/Atom">H1 Collaboration</name> </author> <arxiv:comment xmlns:arxiv="http://arxiv.org/schemas/atom">23 pages, 8 figures and 4 tables</arxiv:comment> <arxiv:journal_ref xmlns:arxiv="http://arxiv.org/schemas/atom">Eur.Phys.J. C31 (2003) 17-29</arxiv:journal_ref> <link xmlns="http://www.w3.org/2005/Atom" href="http://arxiv.org/abs/hep-ex/0307015v1" rel="alternate" type="text/html"/> <link xmlns="http://www.w3.org/2005/Atom" title="pdf" href="http://arxiv.org/pdf/hep-ex/0307015v1" rel="related" type="application/pdf"/> <arxiv:primary_category xmlns:arxiv="http://arxiv.org/schemas/atom" term="hep-ex" scheme="http://arxiv.org/schemas/atom"/> <category term="hep-ex" scheme="http://arxiv.org/schemas/atom"/> </entry> </feed>
```

#### 4.1.1. Perl

[LWP](http://search.cpan.org/%3Csub%3Egaas/libwww-perl-5.808/lib/LWP.pm) is in the default perl installation on most platforms. It can be downloaded and installed from [CPAN](http://search.cpan.org/%3C/sub%3Egaas/libwww-perl-5.808/lib/LWP.pm). Sample code to produce the above output is:

```perl
use LWP; use strict; my $url = 'http://export.arxiv.org/api/query?search_query=all:electron&start=0&max_results=1'; my $browser = LWP::UserAgent->new(); my $response = $browser->get($url); print $response->content();
```

#### 4.1.2. Python

The [urllib](http://docs.python.org/lib/module-urllib.html) module is part of the [python standard library](http://docs.python.org/lib/lib.html), and is included in any default installation of python. Sample code to produce the above output in Python 2.7 is:

```python
import urllib url = 'http://export.arxiv.org/api/query?search_query=all:electron&start=0&max_results=1' data = urllib.urlopen(url).read() print data
```

wheras in Python 3 an example would be:

```python
import urllib.request as libreq with libreq.urlopen('http://export.arxiv.org/api/query?search_query=all:electron&start=0&max_results=1') as url: r = url.read() print(r)
```

#### 4.1.3. Ruby

The [net/http](http://www.ruby-doc.org/stdlib/libdoc/net/http/rdoc/index.html) and [uri](http://www.ruby-doc.org/stdlib/libdoc/uri/rdoc/) modules are part of the [ruby standard library](http://www.ruby-doc.org/stdlib/), and are included in any default installation of ruby. Sample code to produce the above output is:

```ruby
require 'net/http' require 'uri' url = URI.parse('http://export.arxiv.org/api/query?search_query=all:electron&start=0&max_results=1') res = Net::HTTP.get_response(url) print res.body
```

#### 4.1.4. PHP

The file\_get\_contents() function is part of the PHP core language:

```php
<?php $url = 'http://export.arxiv.org/api/query?search_query=all:electron&start=0&max_results=1'; $response = file_get_contents($url); print_r($response); ?>
```

### 4.2. Detailed Parsing Examples

The examples above don't cover how to parse the Atom results returned to extract the information you might be interested in. They also don't cover how to do more advanced programming of the API to perform such tasks as downloading chunks of the full results list one page at a time. The table below contains links to more detailed examples for each of the languages above, as well as to the libraries used to parse Atom.

## 5\. Appendices

### 5.1. Details of Query Construction

As outlined in the [Structure of the API](https://info.arxiv.org/help/api/user-manual.html#Architecture) section, the interface to the API is quite simple. This simplicity, combined with `search_query` construction, and result set filtering through `id_list` makes the API a powerful tool for harvesting data from the arXiv. In this section, we outline the possibilities for constructing `search_query`'s to retrieve our desired article lists. We outlined how to use the `id_list` parameter to filter results sets in [search\_query and id\_list logic](https://info.arxiv.org/help/api/user-manual.html#search_query_and_id_list).

In the arXiv search engine, each article is divided up into a number of fields that can individually be searched. For example, the titles of an article can be searched, as well as the author list, abstracts, comments and journal reference. To search one of these fields, we simply prepend the field prefix followed by a colon to our search term. For example, suppose we wanted to find all articles by the author `Adrian Del Maestro`. We could construct the following query

[http://export.arxiv.org/api/query?search\_query=au:del\_maestro](http://export.arxiv.org/api/query?search_query=au:del_maestro)

This returns nine results. The following table lists the field prefixes for all the fields that can be searched.

<sup>Note: The <code>id_list</code> parameter should be used rather than <code>search_query=id:xxx</code> to properly handle article versions. In addition, note that <code>all:</code> searches in each of the fields simultaneously.</sup>

The API provides one date filter, `submittedDate`, that allow you to select data within a given date range of when the data was submitted to arXiv. The expected format is `[YYYYMMDDTTTT+TO+YYYYMMDDTTTT]` were the `TTTT` is provided in 24 hour time to the minute, in GMT. We could construct the following query using `submittedDate`.

[https://export.arxiv.org/api/query?search\_query=au:del\_maestro+AND+submittedDate:\[202301010600+TO+202401010600\]](https://export.arxiv.org/api/query?search_query=au:del_maestro+AND+submittedDate:[202301010600+TO+202401010600])

The API allows advanced query construction by combining these search fields with Boolean operators. For example, suppose we want to find all articles by the author `Adrian DelMaestro` that also contain the word `checkerboard` in the title. We could construct the following query, using the `AND` operator:

[http://export.arxiv.org/api/query?search\_query=au:del\_maestro+AND+ti:checkerboard](http://export.arxiv.org/api/query?search_query=au:del_maestro+AND+ti:checkerboard)

As expected, this query picked out the one of the nine previous results with `checkerboard` in the title. Note that we included `+` signs in the urls to the API. In a url, a `+` sign encodes a space, which is useful since spaces are not allowed in url's. It is always a good idea to escape the characters in your url's, which is a common feature in most programming libraries that deal with url's. Note that the `<title>` of the returned feed has spaces in the query constructed. It is a good idea to look at `<title>` to see if you have escaped your url correctly.

The following table lists the three possible Boolean operators.

The `ANDNOT` Boolean operator is particularly useful, as it allows us to filter search results based on certain fields. For example, if we wanted all of the articles by the author `Adrian DelMaestro` with titles that _did not_ contain the word `checkerboard`, we could construct the following query:

[http://export.arxiv.org/api/query?search\_query=au:del\_maestro+ANDNOT+ti:checkerboard](http://export.arxiv.org/api/query?search_query=au:del_maestro+ANDNOT+ti:checkerboard)

As expected, this query returns eight results.

Finally, even more complex queries can be used by using parentheses for grouping the Boolean expressions. To include parentheses in in a url, use `%28` for a left-parens `(`, and `%29` for a right-parens `)`. For example, if we wanted all of the articles by the author `Adrian DelMaestro` with titles that _did not_ contain the words `checkerboard`, OR `Pyrochore`, we could construct the following query:

[http://export.arxiv.org/api/query?search\_query=au:del\_maestro+ANDNOT+%28ti:checkerboard+OR+ti:Pyrochlore%29](http://export.arxiv.org/api/query?search_query=au:del_maestro+ANDNOT+%28ti:checkerboard+OR+ti:Pyrochlore%29)

This query returns three results. Notice that the `<title>` element displays the parenthesis correctly meaning that we used the correct url escaping.

So far we have only used single words as the field terms to search for. You can include entire phrases by enclosing the phrase in double quotes, escaped by `%22`. For example, if we wanted all of the articles by the author `Adrian DelMaestro` with titles that contain `quantum criticality`, we could construct the following query:

[http://export.arxiv.org/api/query?search\_query=au:del\_maestro+AND+ti:%22quantum+criticality%22](http://export.arxiv.org/api/query?search_query=au:del_maestro+AND+ti:%22quantum+criticality%22)

This query returns one result, and notice that the feed `<title>` contains double quotes as expected. The table below lists the two grouping operators used in the API.

#### 5.1.1. A Note on Article Versions

Each arXiv article has a version associated with it. The first time an article is posted, it is given a version number of 1. When subsequent corrections are made to an article, it is resubmitted, and the version number is incremented. At any time, any version of an article may be retrieved.

When using the API, if you want to retrieve the latest version of an article, you may simply enter the arxiv id in the `id_list` parameter. If you want to retrieve information about a specific version, you can do this by appending `vn` to the id, where `n` is the version number you are interested in.

For example, to retrieve the latest version of `cond-mat/0207270`, you could use the query [http://export.arxiv.org/api/query?id\_list=cond-mat/0207270](http://export.arxiv.org/api/query?id_list=cond-mat/0207270). To retrieve the very first version of this article, you could use the query [http://export.arxiv.org/api/query?id\_list=cond-mat/0207270v1](http://export.arxiv.org/api/query?id_list=cond-mat/0207270v1)

### 5.2. Details of Atom Results Returned

The following table lists each element of the returned Atom results. For a more detailed explanation see [Outline of an Atom Feed](https://info.arxiv.org/help/api/user-manual.html#atom_feed_outline).

### 5.3. Subject Classifications

For the complete list of arXiv subject classifications, please visit the [taxonomy](https://arxiv.org/category_taxonomy) page.