import * as fs from 'node:fs'
import {SummarizerManager} from 'npm:node-summarizer'

function readingReadMD(file:string) : string {
    const read:string = fs.readFileSync(file,"utf-8");
    const content =  read.split(/\r?\n/)
    let content_str:string = ""
    for(let c of content) {
        content_str += c+"\n";
    }
    return content_str;
} 
function Summarize(readme_line:string) : void {
    const summarizeobject = new SummarizerManager(readme_line,50); // inititally setting to 50
    const sum = summarizeobject.getSummaryByFrequency();
    const file:string = "./context.txt";
    const sentences = sum.summary.split('. '); 
    const final = sentences.join('.\n');
    fs.writeFileSync(file,final);
}

const readme:string =  "./README.md"; // The convention for naming a readme is READMe.md so that is the file that we need to look 
// parent class and get the summary of readme.md.
const temp:string = readingReadMD(readme);
Summarize(temp);
