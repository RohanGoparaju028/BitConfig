<a name="ReadmeSummarizer"></a>

## ReadmeSummarizer
Readme Summarizer

**Kind**: global class  
**Author**: Benny Megidish  

* [ReadmeSummarizer](#ReadmeSummarizer)
    * [.fromUrl(url, short)](#ReadmeSummarizer.fromUrl) ⇒ <code>Promise.&lt;string&gt;</code>
    * [.fromMarkdownText(text, short)](#ReadmeSummarizer.fromMarkdownText) ⇒ <code>Promise.&lt;string&gt;</code>

<a name="ReadmeSummarizer.fromUrl"></a>

### ReadmeSummarizer.fromUrl(url, short) ⇒ <code>Promise.&lt;string&gt;</code>
Summarize markdown readme file from a given URL

**Kind**: static method of [<code>ReadmeSummarizer</code>](#ReadmeSummarizer)  
**Returns**: <code>Promise.&lt;string&gt;</code> - a promise to receive the readme summary.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| url | <code>string</code> |  | readme file URL. |
| short | <code>boolean</code> | <code>true</code> | get short summary (default: true). |

<a name="ReadmeSummarizer.fromMarkdownText"></a>

### ReadmeSummarizer.fromMarkdownText(text, short) ⇒ <code>Promise.&lt;string&gt;</code>
Summarize markdown readme file from a given text string

**Kind**: static method of [<code>ReadmeSummarizer</code>](#ReadmeSummarizer)  
**Returns**: <code>Promise.&lt;string&gt;</code> - a promise to receive the readme summary.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| text | <code>string</code> |  | readme file as string. |
| short | <code>boolean</code> | <code>true</code> | get short summary (default: true). |

