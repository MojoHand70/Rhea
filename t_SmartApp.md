I like it 
- Business Pack is a collection of Rhea's Smart Apps (short: an App)
- Full name: Rhea for Accounting Offices Business pack. 
Now, I think we should go through what an App is. How it works. What it consists of. How it is designed. How the design is presented. What syntax it uses. 
I think we have to put it clear. 
- An App is a program which manipulates data 
- An App knows how to calculate
- An App can run through a defined flow of states but doesn't have to. An App can be stateless and 90% of apps will be so. 
- An app can have a visualisation. The simplest app visualisation is probably a table. What we already have. Every TAB in face is an App. 
- So, to summarize, App definition consists of 
  -- variables
  -- formulas (loops through collections only)
  -- visualization (interactive, can depend on variables)
  -- worflow defined by states, transitions between states, actions performed on transitions, user rights
  -- data structures (an App can define its own data structures, eg. App to create a Client)
  -- external data structures (an App can use data from another App or Rhea core)
  -- ability to trigger connectors (eg. API Fetch, send email, etc...)
- App should respect other object security
- App should define who can run it
- Rhea needs to know how to design an app in an interactive way with user (sandbox, visual artefact) 
In a matter of fact, an App is a small programm run in Rhea. 
I would say, that every tab we have developed till now is an App. For example, let's take a random app from current Rhea UI: "Terminy Podatkowe", app definition: 
{
category: Today
Folder: ''
display: table, default columns: a, b, c, editable=no 
table definition: source: <I don't know, where we keep it, but this can be just a query> (some table properties options like 
workflow: ''
rigths: view: ALL
}
- Definitions has to be light and nimble in order not to impact performance
- performance is of upmost importance

What would you say? We use AI to build programs inside Rhea. Do you think it's possible such approach?