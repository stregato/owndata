
import java.io.IOException;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;


class Identity {

    @JsonProperty("i")
    public String id;
    @JsonProperty("p")
    public String privat;

    public Identity(String nick) throws IOException {
        var m = StashLibrary.instance.stash_newIdentity(nick).map(String.class, String.class);
        id =  m.get("i");
        privat = m.get("p");
    }

    public String  toJson() throws JsonProcessingException {
        return new ObjectMapper().writeValueAsString(this);
    }

}
