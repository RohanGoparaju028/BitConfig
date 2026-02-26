BitConfig is developed with the aim of a command line tool that tracks the programming project that is being developed suck as what are the languages that are used and what are the dependencies that the project is using right now and store that information as a json along with the number of lines that the dependecy file have so when we add more dependencies we can just do the difference of current lenght to already stored lenght.

The main purpose of this cli tool is to get context of the project,which is done by reading the README.md markdown file and summarizing it and asks the developer to give additional context and stores in a text format and push the context that it got from readme summarization and the context that we provided to the large language models(llm) cli's to get better results in the areas where we stuck solving.

The core features that we as a BitConfig team are developing are 
   1) init to initialize the .BitConfig file which creates what programming tech stack we are using 
   2) help to give the basic syntax
   3) get-context to summarize the README.md and prompts the user to enter additinal context that we are doing in the project or we could give the query we need llms to solve the doubts 
   4) push-context pushes the context to the llms and llm provides the results for the project
   5) update updates the .BitConfig  by adding new dependencies that we added sinces the initializing the .BitConfig or last update
   6) diff tells the list of dependencies that are not present in .BitConfig but present in your project.
we are intended to add more feature as the project grows.

Right now we have implemented init help and get-context features and to try the implemented features clone this repo and buil the program using go build -o command and you can execute this feature for yourself.

The resultant binaries for MacOS and Windows is different 
MacOS:- the binary is implemented by ./BuildName 
Windows:- try using go build -o BuildName.exe main.go and type BuildName.exe <command>
