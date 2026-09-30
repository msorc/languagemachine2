# Language Machine: makefile for examples of the language machine in action
# (C) Copyright Peri Hankey 2005 - mpah@users.sourceforge.net - license Gnu GPL
# this file comes with absolutely no warranty




include ../config/config.mak
 
WIKI2HTML = ./makesite.lm
 

HTMLFILES=  mediawiki.html
HTMLFILES+= sitehtml.html
HTMLFILES+= wiki2make.html
HTMLFILES+= testmakefile.html

all: wiki2make.lm testmakefile.mak makesite.lm $(HTMLFILES)
	 
clean:
	rm -f *.html
	rm -f *.lm *.out *.lmr *.o *~  
	rm -f *.tcc *.tCc *.gcc *.gCc *.gdc *.gDc 
	 
makesite.lm: mediawiki.lmn sitehtml.lmn
	$(LMN2M) $+ -o $@ -s $(LM); 
	chmod +x $@
 
# rule to create .mak from .wiki - makefile with mediawiki annotation
%.mak: %.wiki
	./wiki2make.lm $+ -o $*.mak
 
# fake dependency rule to force conversion to .mak
%.2mak: %.wiki
	./wiki2make.lm $+ -o $*.mak
 
# rule to create .html from .lmn - metalanguage source with mediawiki annotation
%.html : %.lmn
	$(WIKI2HTML) $+ -o $*.html 
 
# fake dependency rule to force conversion .html from .lmn
%.2html : %.lmn
	$(WIKI2HTML) $+ -o $*.html 
 
# rule to create .html from .wiki - mediawiki text files
%.html : %.wiki
	$(WIKI2HTML) $+ -o $*.html 
 
# fake dependency rule to force conversion to .html from .wiki
%.2html : %.wiki
	$(WIKI2HTML) $+ -o $*.html 

# make object file from d source
%.o : %.d
	$(GDC) $(DFLAGS) -c -o $@ $< 
 
%.so : %.d
	$(GDC) $(DFLAGS) -c -shared -fPIC -o $@ $<
 
# compile rules to lm loader format
%.lmr : %.lmn
	$(LMN2M) -o $@ $+

# make shebang script from rules: extensions have to be registered with lm engine
%.lm :  %.lmn
	$(LMN2M) -o $@ -s $(LM) $+
	chmod +x $@
 
# make D wrapper for lmn rules: extensions are functions with C linkdage
%.2d : %.lmn
	$(LMN2D) -o  $*.d $<
 
# compile lmn rules to D code:  extensions are functions with C linkdage
%.2D : %.lmn
	$(LMN2DD) -o $*.d $<
 
# make C wrapper for lmn rules: extensions are functions with C linkdage
%.2c : %.lmn
	$(LMN2C) -o  $*.c $<
 
# compile lmn rules to C code:  extensions are functions with C linkdage
%.2C : %.lmn
	$(LMN2CC) -o $*.c $<
 
# compile and link lmn rules as C code with no extensions using gcc
%.gcc : %.lmn
	$(LMN2C)  -o $*.c -c $<
	gcc -o $@  $*.c -ldl $(LMLIB)/liblm -Wl,-rpath,$(LMLIB)/ -L$(GDCPREFIX)/lib/ -lgphobos -lm -lpthread -lc
 
%.gCc : %.lmn
	$(LMN2CC) -o $*.c -c $<
	gcc -o $@  $*.c -ldl $(LMLIB)/liblm -Wl,-rpath,$(LMLIB)/ -L$(GDCPREFIX)/lib/ -lgphobos -lm -lpthread -lc
 
# compile and link lmn rules as D code with no extensions using gdc
%.gdc : %.lmn
	$(LMN2D)  -o $*.d -d $<
	$(GDC) -o $@ $(DFLAGS) $*.d -ldl $(LMLIB)/liblm -Wl,-rpath,$(LMLIB)/ 
 
%.gDc : %.lmn
	$(LMN2DD) -o $*.d -d $<
	$(GDC) -o $@ $(DFLAGS) $*.d -ldl $(LMLIB)/liblm -Wl,-rpath,$(LMLIB)/ 
