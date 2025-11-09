`timescale 10ns/100ps

module hello_test();

   reg [6:0] tick;
   reg	     reset;
   wire	     red, green, blue;

   top #(.TOP_BIT(2)) top(reset, tick[1], red, green, blue);

   initial
     begin
        $dumpvars();
	tick <= #1 0;
	reset <= #1 1;
     end

   always @(tick)
     begin
	case (tick)
	  2:
	    begin
	       reset <= #1 0;
	    end
	  7:
	    begin
	       reset <= #1 1;
	    end
	  40:
	    begin
	       $display("end of test");
	       $finish_and_return(0);
	    end
	endcase
	tick <= #1 (tick + 1);
     end

endmodule // hello_test
